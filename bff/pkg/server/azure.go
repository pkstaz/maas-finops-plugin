package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type azurePriceCache struct {
	mu      sync.Mutex
	expires time.Time
	values  map[string]CostSource
}

var azureCache = &azurePriceCache{values: map[string]CostSource{}}

type azureRetailPage struct {
	Items []struct {
		RetailPrice   float64 `json:"retailPrice"`
		ArmSkuName    string  `json:"armSkuName"`
		ArmRegion     string  `json:"armRegionName"`
		SkuName       string  `json:"skuName"`
		MeterName     string  `json:"meterName"`
		ProductName   string  `json:"productName"`
		Type          string  `json:"type"`
		UnitOfMeasure string  `json:"unitOfMeasure"`
	} `json:"Items"`
}

func fetchAzureRetail(sku, region string) (CostSource, bool) {
	if region == "" {
		region = "eastus"
	}
	key := strings.ToLower(sku + "|" + region)
	azureCache.mu.Lock()
	if time.Now().Before(azureCache.expires) {
		if v, ok := azureCache.values[key]; ok {
			azureCache.mu.Unlock()
			return v, v.HourlyUSD > 0
		}
	}
	azureCache.mu.Unlock()

	filter := fmt.Sprintf("armSkuName eq '%s' and armRegionName eq '%s' and priceType eq 'Consumption'", sku, region)
	u := "https://prices.azure.com/api/retail/prices?$filter=" + url.QueryEscape(filter)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return CostSource{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return CostSource{}, false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CostSource{}, false
	}
	var page azureRetailPage
	if json.Unmarshal(body, &page) != nil {
		return CostSource{}, false
	}

	var picked *CostSource
	for i := range page.Items {
		it := page.Items[i]
		name := strings.ToLower(it.SkuName + " " + it.MeterName + " " + it.ProductName)
		if strings.Contains(name, "spot") || strings.Contains(name, "low priority") || strings.Contains(name, "windows") {
			continue
		}
		if it.RetailPrice <= 0 {
			continue
		}
		cs := CostSource{
			HourlyUSD: it.RetailPrice,
			Kind:      "azure-retail",
			SKU:       sku,
			Region:    region,
			Meter:     it.MeterName,
			Note:      "Azure Retail Prices PAYG Linux",
		}
		picked = &cs
		if !strings.Contains(strings.ToLower(it.SkuName), "spot") {
			break
		}
	}
	if picked == nil {
		return CostSource{}, false
	}
	azureCache.mu.Lock()
	azureCache.values[key] = *picked
	azureCache.expires = time.Now().Add(6 * time.Hour)
	azureCache.mu.Unlock()
	return *picked, true
}
