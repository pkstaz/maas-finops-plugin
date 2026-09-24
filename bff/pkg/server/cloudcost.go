package server

import "strings"

func lookupMachineCost(hw HardwareConfig, sku, region, provider string) CostSource {
	sku = strings.TrimSpace(sku)
	region = strings.ToLower(strings.TrimSpace(region))
	provider = normalizeProvider(firstNonEmpty(provider, hw.Provider, inferProviderFromSKU(sku)))
	if sku == "" {
		return CostSource{Kind: "missing", Note: "no machine SKU"}
	}

	if cost, ok := manualMachineCost(hw, sku, region); ok {
		return cost
	}

	switch provider {
	case "azure":
		if cost, ok := fetchAzureRetail(sku, region); ok {
			return cost
		}
		if p, ok := catalogPrice("azure", sku); ok {
			return CostSource{
				HourlyUSD: p, Kind: "catalog", SKU: sku,
				Region: firstNonEmpty(region, "eastus"),
				Note:   "local Azure PAYG Linux catalog (retail API unavailable)",
			}
		}
	case "aws":
		if p, ok := catalogPrice("aws", sku); ok {
			return CostSource{
				HourlyUSD: p, Kind: "catalog", SKU: sku,
				Region: firstNonEmpty(region, "us-east-1"),
				Note:   "AWS Linux on-demand catalog (us-east-1 list); reservations/EDP manually",
			}
		}
	case "ibmcloud":
		if p, ok := catalogPrice("ibmcloud", sku); ok {
			return CostSource{
				HourlyUSD: p, Kind: "catalog", SKU: sku,
				Region: firstNonEmpty(region, "us-south"),
				Note:   "IBM Cloud VPC GPU catalog (us-south list); confirm in your account or enter manually",
			}
		}
	}

	if s, p, ok := findSKUAny(sku); ok && s.HourlyUSD > 0 {
		return CostSource{
			HourlyUSD: s.HourlyUSD, Kind: "catalog", SKU: sku, Region: firstNonEmpty(region, s.Region),
			Note: p + " catalog",
		}
	}

	return CostSource{
		Kind:   "missing",
		SKU:    sku,
		Region: region,
		Note:   "no list price for " + sku + " on " + firstNonEmpty(provider, "this cloud") + "; enter USD/hour manually",
	}
}

func manualMachineCost(hw HardwareConfig, sku, region string) (CostSource, bool) {
	for _, m := range hw.Machines {
		if !strings.EqualFold(m.InstanceType, sku) {
			continue
		}
		if m.Region != "" && region != "" && !strings.EqualFold(m.Region, region) {
			continue
		}
		if m.HourlyUSD <= 0 {
			continue
		}
		return CostSource{
			HourlyUSD: m.HourlyUSD,
			Kind:      "manual",
			SKU:       sku,
			Region:    firstNonEmpty(m.Region, region),
			Note:      firstNonEmpty(m.Notes, "manually entered cost"),
		}, true
	}
	return CostSource{}, false
}

func catalogPrice(provider, sku string) (float64, bool) {
	s, ok := findSKU(provider, sku)
	if !ok || s.HourlyUSD <= 0 {
		return 0, false
	}
	return s.HourlyUSD, true
}

func normalizeProvider(platform string) string {
	p := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(platform), " ", ""))
	p = strings.ReplaceAll(p, "_", "")
	p = strings.ReplaceAll(p, "-", "")
	switch p {
	case "azure":
		return "azure"
	case "aws", "amazon", "amazonwebservices":
		return "aws"
	case "ibmcloud", "ibm", "roks":
		return "ibmcloud"
	case "baremetal", "none", "metal":
		return "baremetal"
	case "auto", "unknown", "":
		return ""
	default:
		return p
	}
}

func inferProviderFromSKU(sku string) string {
	s := strings.ToLower(sku)
	switch {
	case strings.HasPrefix(s, "standard_"):
		return "azure"
	case strings.HasPrefix(s, "gx2-"), strings.HasPrefix(s, "gx3"):
		return "ibmcloud"
	case strings.Contains(s, "."):
		return "aws"
	default:
		return ""
	}
}

func providerSupported(provider string) bool {
	switch normalizeProvider(provider) {
	case "azure", "aws", "ibmcloud", "baremetal":
		return true
	default:
		return false
	}
}
