package server

import (
	"fmt"
	"math"
	"strings"
)

func normalizeHardware(h HardwareConfig) HardwareConfig {
	if h.Utilization <= 0 || h.Utilization > 1 {
		h.Utilization = 0.8
	}
	if h.Margin < 0 {
		h.Margin = 0
	}
	if h.PriceType == "" {
		h.PriceType = "Consumption"
	}
	if h.Machines == nil {
		h.Machines = []ManualMachine{}
	}
	return h
}

func recommendPrices(inv ClusterInventory, cat PricingCatalog) RecommendResponse {
	hw := normalizeHardware(cat.Hardware)
	if hw.Provider == "" || strings.EqualFold(hw.Provider, "auto") {
		hw.Provider = inv.Provider
	}
	out := RecommendResponse{
		Inventory:   inv,
		Utilization: hw.Utilization,
		Margin:      hw.Margin,
		Currency:    firstNonEmpty(cat.Currency, "USD"),
		Models:      []ModelRecommendation{},
		Notes: []string{
			fmt.Sprintf("Capacity used: %.0f%% of estimated GPU throughput (not 100%% of the card).", hw.Utilization*100),
			costNoteForProvider(inv.Provider),
			"RHOAI-hosted models (LLMInferenceService): token price from GPU USD/hour.",
			"External models: vendor token list (Azure OpenAI, OpenAI, …) or a price you enter. Cluster GPU cost does not apply.",
		},
	}
	if !inv.Supported && len(hw.Machines) == 0 {
		out.NeedsManual = true
		out.Notes = append(out.Notes, firstNonEmpty(inv.Message, "No list price for this provider: add GPU hourly cost manually."))
	}

	byName := map[string]NodeInventory{}
	var gpuNode NodeInventory
	for _, n := range inv.Nodes {
		byName[n.Name] = n
		if n.GPUCount > 0 && gpuNode.Name == "" {
			gpuNode = n
		}
	}

	models := inv.Models
	var rhoaiCount, externalCount int
	for _, m := range models {
		if isExternalOrigin(m.Origin, m.Kind) {
			externalCount++
		} else {
			rhoaiCount++
		}
	}
	if rhoaiCount == 0 && externalCount == 0 {
		out.Notes = append(out.Notes, "No MaaS models found (LLMInferenceService or ExternalModel).")
	} else if rhoaiCount == 0 {
		out.Notes = append(out.Notes, "No RHOAI-hosted models. External models use vendor token rates, not GPU cost.")
	}

	for _, m := range models {
		var rec ModelRecommendation
		if isExternalOrigin(m.Origin, m.Kind) {
			rec = buildExternalRecommendation(m, cat, hw)
		} else {
			node := gpuNode
			if n, ok := byName[m.NodeName]; ok {
				node = n
			}
			provider := firstNonEmpty(inv.Provider, hw.Provider)
			rec = buildModelRecommendation(m, node, hw, provider)
		}
		if rec.Cost.Kind == "missing" {
			out.NeedsManual = true
		}
		out.Models = append(out.Models, rec)
	}
	return out
}

func buildModelRecommendation(m ServingModel, node NodeInventory, hw HardwareConfig, provider string) ModelRecommendation {
	sku := node.InstanceType
	region := firstNonEmpty(node.Region, hwRegionFallback(provider))
	gpu := firstNonEmpty(node.GPUProduct, "unknown")
	cost := lookupMachineCost(hw, sku, region, provider)

	tokFull := estimateTokensPerSec(gpu, m.ParamsB, m.Quant, node.GPUCount)
	tokUsed := tokFull * hw.Utilization
	tokHour := tokUsed * 3600
	hourly := cost.HourlyUSD * (1 + hw.Margin)

	rec := ModelRecommendation{
		Name:             m.Name,
		DisplayName:      m.DisplayName,
		Namespace:        m.Namespace,
		Origin:           firstNonEmpty(m.Origin, originRHOAI),
		Kind:             firstNonEmpty(m.Kind, "LLMInferenceService"),
		InstanceType:     sku,
		GPUProduct:       gpu,
		Quant:            m.Quant,
		ParamsB:          m.ParamsB,
		TokensPerSecFull: round4(tokFull),
		TokensPerSecUsed: round4(tokUsed),
		TokensPerHour:    round4(tokHour),
		HourlyUSD:        round4(hourly),
		Cost:             cost,
		CanApply:         cost.HourlyUSD > 0 && tokHour > 0,
	}

	if tokHour > 0 && hourly > 0 {
		rec.PricePerMillion = round4((hourly / tokHour) * 1_000_000)
	}

	rec.Assumptions = []string{
		fmt.Sprintf("Machine %s in %s", firstNonEmpty(sku, "?"), firstNonEmpty(region, "?")),
		fmt.Sprintf("GPU %s ×%d, model ~%.1fB %s", gpu, max(node.GPUCount, 1), m.ParamsB, m.Quant),
		fmt.Sprintf("%.0f tok/s at 100%% → %.0f tok/s at %.0f%%", tokFull, tokUsed, hw.Utilization*100),
		fmt.Sprintf("USD/hour %.4f (%s) / tokens-hour %.0f × 1e6", hourly, cost.Kind, tokHour),
	}
	if m.MaxModelLen > 0 {
		rec.Assumptions = append(rec.Assumptions, fmt.Sprintf("max-model-len=%d (context; does not change base $/token)", m.MaxModelLen))
	}
	if cost.Kind == "missing" {
		rec.Warning = "Add hourly cost for " + firstNonEmpty(sku, "this GPU") + " under Manual machine cost."
	}
	if tokFull <= 0 {
		rec.Warning = "No credible throughput on this GPU (VRAM / size). Set price manually."
		rec.CanApply = false
	}
	return rec
}

func buildExternalRecommendation(m ServingModel, cat PricingCatalog, hw HardwareConfig) ModelRecommendation {
	rec := ModelRecommendation{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Namespace:   m.Namespace,
		Origin:      originExternal,
		Kind:        firstNonEmpty(m.Kind, "ExternalModel"),
		Provider:    m.Provider,
		Endpoint:    m.Endpoint,
		Cost: CostSource{
			Kind: "missing",
			SKU:  firstNonEmpty(m.TargetModel, m.Name),
			Note: "external model — enter USD / 1M tokens (GPU hourly cost does not apply)",
		},
	}
	if mp, ok := cat.Models[m.Name]; ok && (mp.PricePerMillion > 0 || mp.InputPerMillion > 0 || mp.OutputPerMillion > 0) {
		rec.InputPerMillion = mp.InputPerMillion
		rec.OutputPerMillion = mp.OutputPerMillion
		rec.PricePerMillion = mp.PricePerMillion
		if rec.PricePerMillion <= 0 {
			rec.PricePerMillion = blendedPerMillion(mp.InputPerMillion, mp.OutputPerMillion)
		}
		rec.Cost = CostSource{Kind: firstNonEmpty(mp.Source, "catalog"), SKU: m.Name, Note: "saved catalog price"}
		rec.CanApply = false
		rec.Assumptions = []string{
			fmt.Sprintf("External %s via %s", firstNonEmpty(m.TargetModel, m.Name), firstNonEmpty(m.Provider, "provider")),
			firstNonEmpty(m.Endpoint, "endpoint unknown"),
			"Using the price already saved in the FinOps catalog.",
		}
		return rec
	}
	if v, ok := lookupVendor(m.Provider, firstNonEmpty(m.TargetModel, m.Name)); ok {
		in := v.InputPerMillion * (1 + hw.Margin)
		out := v.OutputPerMillion * (1 + hw.Margin)
		rec.InputPerMillion = round4(in)
		rec.OutputPerMillion = round4(out)
		rec.PricePerMillion = blendedPerMillion(in, out)
		rec.Cost = CostSource{
			Kind: "vendor", SKU: v.Model, Region: m.Provider,
			Note: v.Note + "; confirm in your account or enter manually",
		}
		rec.CanApply = rec.PricePerMillion > 0
		rec.Assumptions = []string{
			fmt.Sprintf("External %s (%s)", v.Model, v.Provider),
			firstNonEmpty(m.Endpoint, "endpoint unknown"),
			fmt.Sprintf("List USD / 1M: input %.4f · output %.4f", in, out),
			"Blended / 1M assumes 3:1 input:output when the gateway only reports total tokens.",
		}
		return rec
	}
	rec.Warning = "No vendor list price for this external model. Enter USD / 1M (input/output) on Pricing."
	rec.Assumptions = []string{
		fmt.Sprintf("External %s via %s", firstNonEmpty(m.TargetModel, m.Name), firstNonEmpty(m.Provider, "provider")),
		firstNonEmpty(m.Endpoint, "endpoint unknown"),
		"Not hosted on cluster GPU — do not use the hardware token formula.",
	}
	return rec
}

func estimateTokensPerSec(gpu string, paramsB float64, quant string, gpuCount int) float64 {
	if paramsB <= 0 {
		paramsB = 8
	}
	g := strings.ToLower(gpu)
	var fp16 float64
	switch {
	case strings.Contains(g, "h200"):
		fp16 = 1100
	case strings.Contains(g, "h100"):
		fp16 = 900
	case strings.Contains(g, "a100"):
		fp16 = 220
	case strings.Contains(g, "l40"):
		fp16 = 180
	case strings.Contains(g, "l4"):
		fp16 = 80
	case strings.Contains(g, "a10"):
		fp16 = 90
	case strings.Contains(g, "v100"):
		fp16 = 80
	case strings.Contains(g, "t4") || strings.Contains(g, "tesla t4"):
		fp16 = 40
	default:
		fp16 = 40
	}

	mult := 1.0
	switch quant {
	case "w4a16", "int4", "awq", "gptq":
		mult = 2.0
	case "w8a8", "int8", "fp8":
		mult = 1.5
	case "bf16", "fp16":
		mult = 1.0
	default:
		mult = 1.7
	}

	// 8B reference; scale roughly inverse with params.
	scale := 8.0 / paramsB
	if scale > 4 {
		scale = 4
	}
	if scale < 0.15 {
		scale = 0.15
	}
	n := gpuCount
	if n < 1 {
		n = 1
	}
	return fp16 * mult * scale * float64(n)
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func costNoteForProvider(provider string) string {
	switch normalizeProvider(provider) {
	case "aws":
		return "Machine cost = AWS Linux on-demand catalog (us-east-1 list) or the USD/hour you enter (Savings Plans / EDP)."
	case "ibmcloud":
		return "Machine cost = IBM Cloud VPC GPU catalog (us-south list) or your account USD/hour."
	case "azure":
		return "Machine cost = Azure Retail PAYG Linux or the USD/hour you enter (reservations / EA / discounts)."
	case "baremetal":
		return "Machine cost = USD/hour you enter for the GPU server (no public cloud list price)."
	default:
		return "Machine cost = USD/hour entered manually (this cloud has no list price in the plugin)."
	}
}

func hwRegionFallback(provider string) string {
	switch normalizeProvider(provider) {
	case "aws":
		return "us-east-1"
	case "ibmcloud":
		return "us-south"
	case "azure":
		return "eastus"
	default:
		return ""
	}
}
