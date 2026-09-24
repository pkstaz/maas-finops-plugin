package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SimulatorCatalogResponse struct {
	Inventory ClusterInventory  `json:"inventory"`
	Current   SimulatorCurrent  `json:"current"`
	Providers []ProviderCatalog `json:"providers"`
	Models    []CatalogModel    `json:"models"`
	Currency  string            `json:"currency"`
}

type SimulatorCurrent struct {
	Provider     string `json:"provider"`
	SKU          string `json:"sku"`
	Region       string `json:"region"`
	GPUProduct   string `json:"gpuProduct"`
	GPUCount     int    `json:"gpuCount"`
	GPUMemoryGiB int    `json:"gpuMemoryGib"`
	ThisCluster  bool   `json:"thisCluster"`
}

type SimulatorQuote struct {
	Provider         string       `json:"provider"`
	SKU              CatalogSKU   `json:"sku"`
	Model            CatalogModel `json:"model"`
	Utilization      float64      `json:"utilization"`
	TokensPerSecFull float64      `json:"tokensPerSecFull"`
	TokensPerSecUsed float64      `json:"tokensPerSecUsed"`
	TokensPerHour    float64      `json:"tokensPerHour"`
	HourlyUSD        float64      `json:"hourlyUsd"`
	PricePerMillion  float64      `json:"pricePerMillion"`
	VRAMGiB          float64      `json:"vramGib"`
	Fits             bool         `json:"fits"`
	Cost             CostSource   `json:"cost"`
	Assumptions      []string     `json:"assumptions"`
	Warning          string       `json:"warning,omitempty"`
}

func (s *Server) handleSimulatorCatalog(w http.ResponseWriter, r *http.Request) {
	inv := s.clusterInventory(r)
	cur := simulatorCurrent(inv)
	writeJSON(w, http.StatusOK, SimulatorCatalogResponse{
		Inventory: inv,
		Current:   cur,
		Providers: allProviderCatalogs(),
		Models:    redHatModels,
		Currency:  "USD",
	})
}

func (s *Server) handleSimulatorQuote(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	provider := normalizeProvider(q.Get("provider"))
	skuName := q.Get("sku")
	modelID := q.Get("model")
	util := 0.8
	if u := q.Get("utilization"); u != "" {
		if f, err := strconv.ParseFloat(u, 64); err == nil && f > 0 && f <= 1 {
			util = f
		}
	}
	hourlyOverride := 0.0
	if h := q.Get("hourlyUsd"); h != "" {
		hourlyOverride, _ = strconv.ParseFloat(h, 64)
	}
	gpuHints := CatalogSKU{
		GPUProduct: q.Get("gpuProduct"),
	}
	if n, err := strconv.Atoi(q.Get("gpuCount")); err == nil {
		gpuHints.GPUCount = n
	}
	if n, err := strconv.Atoi(q.Get("gpuMemoryGib")); err == nil {
		gpuHints.GPUMemoryGiB = n
	}

	quote, err := simulateQuote(provider, skuName, modelID, util, hourlyOverride, s.pricing.Get(r.Context()), gpuHints)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func simulatorCurrent(inv ClusterInventory) SimulatorCurrent {
	out := SimulatorCurrent{Provider: inv.Provider, Region: inv.Region, ThisCluster: true}
	for _, n := range inv.Nodes {
		if n.GPUCount <= 0 {
			continue
		}
		out.SKU = n.InstanceType
		out.GPUProduct = n.GPUProduct
		out.GPUCount = n.GPUCount
		if n.GPUMemoryMiB > 0 {
			out.GPUMemoryGiB = n.GPUMemoryMiB / 1024
		}
		if out.Region == "" {
			out.Region = n.Region
		}
		break
	}
	return out
}

func simulateQuote(provider, skuName, modelID string, util, hourlyOverride float64, cat PricingCatalog, gpuHints ...CatalogSKU) (SimulatorQuote, error) {
	if provider == "" {
		return SimulatorQuote{}, fmt.Errorf("provider is required")
	}
	if skuName == "" {
		return SimulatorQuote{}, fmt.Errorf("SKU is required")
	}
	model, ok := findCatalogModel(modelID)
	if !ok {
		return SimulatorQuote{}, fmt.Errorf("unknown Red Hat model %q", modelID)
	}
	skuInfo, ok := findSKU(provider, skuName)
	if !ok {
		if any, _, found := findSKUAny(skuName); found {
			skuInfo = any
		} else {
			skuInfo = CatalogSKU{InstanceType: skuName}
		}
	}
	if len(gpuHints) > 0 {
		h := gpuHints[0]
		if h.GPUProduct != "" {
			skuInfo.GPUProduct = h.GPUProduct
		}
		if h.GPUCount > 0 {
			skuInfo.GPUCount = h.GPUCount
		}
		if h.GPUMemoryGiB > 0 {
			skuInfo.GPUMemoryGiB = h.GPUMemoryGiB
		}
	}

	hw := normalizeHardware(cat.Hardware)
	hw.Utilization = util
	cost := lookupMachineCost(hw, skuName, firstNonEmpty(skuInfo.Region, hwRegionFallback(provider)), provider)
	if hourlyOverride > 0 {
		cost.HourlyUSD = hourlyOverride
		cost.Kind = "manual"
		cost.Note = "hourly cost override"
	}
	if cost.HourlyUSD <= 0 && provider == "baremetal" {
		return SimulatorQuote{}, fmt.Errorf("bare metal requires USD/hour")
	}

	gpuCount := skuInfo.GPUCount
	if gpuCount < 1 {
		gpuCount = 1
	}
	tokFull := estimateTokensPerSec(skuInfo.GPUProduct, model.ParamsB, model.Quant, gpuCount)
	tokUsed := tokFull * util
	tokHour := tokUsed * 3600
	hourly := cost.HourlyUSD * (1 + hw.Margin)
	vram := modelVRAMGiB(model.ParamsB, model.Quant)
	totalVRAM := float64(skuInfo.GPUMemoryGiB * gpuCount)
	fits := skuInfo.GPUMemoryGiB == 0 || vram <= totalVRAM*0.9

	out := SimulatorQuote{
		Provider:         provider,
		SKU:              skuInfo,
		Model:            model,
		Utilization:      util,
		TokensPerSecFull: round4(tokFull),
		TokensPerSecUsed: round4(tokUsed),
		TokensPerHour:    round4(tokHour),
		HourlyUSD:        round4(hourly),
		VRAMGiB:          round4(vram),
		Fits:             fits,
		Cost:             cost,
	}
	if tokHour > 0 && hourly > 0 {
		out.PricePerMillion = round4((hourly / tokHour) * 1_000_000)
	}
	out.Assumptions = []string{
		fmt.Sprintf("%s · %s", skuInfo.InstanceType, firstNonEmpty(skuInfo.GPUProduct, "?")),
		fmt.Sprintf("%s · %.1fB %s", model.DisplayName, model.ParamsB, model.Quant),
		fmt.Sprintf("%.0f tok/s at 100%% → %.0f tok/s at %.0f%%", tokFull, tokUsed, util*100),
		fmt.Sprintf("USD/hour %.4f (%s) / tokens-hour %.0f × 1e6", hourly, cost.Kind, tokHour),
		fmt.Sprintf("Estimated weights+KV ~%.1f GiB vs %.0f GiB GPU memory", vram, totalVRAM),
	}
	if cost.HourlyUSD <= 0 {
		out.Warning = "Enter USD/hour for this SKU to get a token price."
	}
	if !fits {
		out.Warning = fmt.Sprintf("This model likely does not fit: ~%.0f GiB needed, SKU has %.0f GiB.", vram, totalVRAM)
	}
	if tokFull <= 0 {
		out.Warning = "No credible throughput for this GPU / model size."
	}
	return out, nil
}

func modelVRAMGiB(paramsB float64, quant string) float64 {
	if paramsB <= 0 {
		paramsB = 8
	}
	bytes := 2.0
	switch strings.ToLower(quant) {
	case "w4a16", "int4", "awq", "gptq":
		bytes = 0.55
	case "w8a8", "int8", "fp8":
		bytes = 1.1
	case "bf16", "fp16":
		bytes = 2.0
	}
	return paramsB * bytes * 1.25
}
