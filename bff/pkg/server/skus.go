package server

import "strings"

type CatalogSKU struct {
	InstanceType string  `json:"instanceType"`
	HourlyUSD    float64 `json:"hourlyUsd"`
	GPUProduct   string  `json:"gpuProduct"`
	GPUCount     int     `json:"gpuCount"`
	GPUMemoryGiB int     `json:"gpuMemoryGib"`
	Region       string  `json:"region,omitempty"`
	Note         string  `json:"note,omitempty"`
}

type ProviderCatalog struct {
	ID    string       `json:"id"`
	Label string       `json:"label"`
	SKUs  []CatalogSKU `json:"skus"`
}

var azureSKUs = []CatalogSKU{
	sku("Standard_NC4as_T4_v3", 0.526, "Tesla T4", 1, 16, "eastus"),
	sku("Standard_NC8as_T4_v3", 0.752, "Tesla T4", 1, 16, "eastus"),
	sku("Standard_NC16as_T4_v3", 1.204, "Tesla T4", 1, 16, "eastus"),
	sku("Standard_NC64as_T4_v3", 4.352, "Tesla T4", 4, 16, "eastus"),
	sku("Standard_NV36ads_A10_v5", 2.01, "A10", 1, 24, "eastus"),
	sku("Standard_NC24ads_A100_v4", 3.673, "A100", 1, 80, "eastus"),
	sku("Standard_NC48ads_A100_v4", 7.346, "A100", 2, 80, "eastus"),
	sku("Standard_NC96ads_A100_v4", 14.692, "A100", 4, 80, "eastus"),
	sku("Standard_ND96isr_H100_v5", 98.32, "H100", 8, 80, "eastus"),
}

var awsSKUs = []CatalogSKU{
	sku("g4dn.xlarge", 0.526, "Tesla T4", 1, 16, "us-east-1"),
	sku("g4dn.2xlarge", 0.752, "Tesla T4", 1, 16, "us-east-1"),
	sku("g4dn.4xlarge", 1.204, "Tesla T4", 1, 16, "us-east-1"),
	sku("g4dn.12xlarge", 3.912, "Tesla T4", 4, 16, "us-east-1"),
	sku("g5.xlarge", 1.006, "A10G", 1, 24, "us-east-1"),
	sku("g5.2xlarge", 1.212, "A10G", 1, 24, "us-east-1"),
	sku("g5.4xlarge", 1.624, "A10G", 1, 24, "us-east-1"),
	sku("g5.12xlarge", 5.672, "A10G", 4, 24, "us-east-1"),
	sku("g6.xlarge", 0.8048, "L4", 1, 24, "us-east-1"),
	sku("g6.2xlarge", 0.9776, "L4", 1, 24, "us-east-1"),
	sku("g6.12xlarge", 4.6016, "L4", 4, 24, "us-east-1"),
	sku("g6e.xlarge", 1.861, "L40S", 1, 48, "us-east-1"),
	sku("g6e.12xlarge", 10.492, "L40S", 4, 48, "us-east-1"),
	sku("p3.2xlarge", 3.06, "V100", 1, 16, "us-east-1"),
	sku("p4d.24xlarge", 32.7726, "A100", 8, 40, "us-east-1"),
	sku("p5.48xlarge", 98.32, "H100", 8, 80, "us-east-1"),
}

var ibmSKUs = []CatalogSKU{
	sku("gx2-8x64x1v100", 2.50, "V100", 1, 16, "us-south"),
	sku("gx2-16x128x1v100", 3.20, "V100", 1, 16, "us-south"),
	sku("gx2-16x128x2v100", 5.00, "V100", 2, 16, "us-south"),
	sku("gx2-32x256x2v100", 6.40, "V100", 2, 16, "us-south"),
	sku("gx3-16x80x1l4", 1.85, "L4", 1, 24, "us-south"),
	sku("gx3-32x160x2l4", 3.70, "L4", 2, 24, "us-south"),
	sku("gx3-64x320x4l4", 7.40, "L4", 4, 24, "us-south"),
	sku("gx3-24x120x1l40s", 3.88, "L40S", 1, 48, "us-south"),
	sku("gx3-48x240x2l40s", 7.76, "L40S", 2, 48, "us-south"),
	sku("gx3d-24x120x1a100p", 5.50, "A100", 1, 80, "us-south"),
	sku("gx3d-48x240x2a100p", 11.00, "A100", 2, 80, "us-south"),
	sku("gx3d-160x1792x8h100", 80.00, "H100", 8, 80, "us-south"),
	sku("gx3d-160x1792x8h200", 95.00, "H200", 8, 141, "us-south"),
}

var baremetalSKUs = []CatalogSKU{
	{InstanceType: "1x Tesla T4", GPUProduct: "Tesla T4", GPUCount: 1, GPUMemoryGiB: 16, Note: "enter USD/hour"},
	{InstanceType: "1x L4", GPUProduct: "L4", GPUCount: 1, GPUMemoryGiB: 24, Note: "enter USD/hour"},
	{InstanceType: "1x A10", GPUProduct: "A10", GPUCount: 1, GPUMemoryGiB: 24, Note: "enter USD/hour"},
	{InstanceType: "1x L40S", GPUProduct: "L40S", GPUCount: 1, GPUMemoryGiB: 48, Note: "enter USD/hour"},
	{InstanceType: "1x A100 80GB", GPUProduct: "A100", GPUCount: 1, GPUMemoryGiB: 80, Note: "enter USD/hour"},
	{InstanceType: "2x A100 80GB", GPUProduct: "A100", GPUCount: 2, GPUMemoryGiB: 80, Note: "enter USD/hour"},
	{InstanceType: "1x H100", GPUProduct: "H100", GPUCount: 1, GPUMemoryGiB: 80, Note: "enter USD/hour"},
	{InstanceType: "8x H100", GPUProduct: "H100", GPUCount: 8, GPUMemoryGiB: 80, Note: "enter USD/hour"},
	{InstanceType: "1x H200", GPUProduct: "H200", GPUCount: 1, GPUMemoryGiB: 141, Note: "enter USD/hour"},
}

func sku(name string, hourly float64, gpu string, count, memGiB int, region string) CatalogSKU {
	return CatalogSKU{
		InstanceType: name, HourlyUSD: hourly, GPUProduct: gpu,
		GPUCount: count, GPUMemoryGiB: memGiB, Region: region,
	}
}

func hardwareCatalog(provider string) []CatalogSKU {
	switch normalizeProvider(provider) {
	case "azure":
		return azureSKUs
	case "aws":
		return awsSKUs
	case "ibmcloud":
		return ibmSKUs
	case "baremetal":
		return baremetalSKUs
	default:
		return nil
	}
}

func allProviderCatalogs() []ProviderCatalog {
	return []ProviderCatalog{
		{ID: "azure", Label: "Azure / ARO", SKUs: azureSKUs},
		{ID: "aws", Label: "AWS / ROSA", SKUs: awsSKUs},
		{ID: "ibmcloud", Label: "IBM Cloud", SKUs: ibmSKUs},
		{ID: "baremetal", Label: "Bare metal", SKUs: baremetalSKUs},
	}
}

func findSKU(provider, name string) (CatalogSKU, bool) {
	for _, s := range hardwareCatalog(provider) {
		if strings.EqualFold(s.InstanceType, name) {
			return s, true
		}
	}
	return CatalogSKU{}, false
}

func findSKUAny(name string) (CatalogSKU, string, bool) {
	for _, p := range []string{"azure", "aws", "ibmcloud", "baremetal"} {
		if s, ok := findSKU(p, name); ok {
			return s, p, true
		}
	}
	return CatalogSKU{}, "", false
}
