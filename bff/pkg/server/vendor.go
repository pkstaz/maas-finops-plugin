package server

import "strings"

const (
	originRHOAI    = "rhoai"
	originExternal = "external"
)

type VendorRate struct {
	Provider         string
	Model            string
	InputPerMillion  float64
	OutputPerMillion float64
	Note             string
}

// Public list USD / 1M tokens. Confirm in the vendor account; EA/PTU overrides go in the catalog.
var vendorRates = []VendorRate{
	{"azure-openai", "gpt-4o", 2.50, 10.00, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4o-mini", 0.15, 0.60, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1", 2.00, 8.00, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1-mini", 0.40, 1.60, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1-nano", 0.10, 0.40, "Azure OpenAI standard list"},
	{"openai", "gpt-4o", 2.50, 10.00, "OpenAI API list"},
	{"openai", "gpt-4o-mini", 0.15, 0.60, "OpenAI API list"},
}

func originFromKind(kind string) string {
	if strings.EqualFold(strings.TrimSpace(kind), "ExternalModel") {
		return originExternal
	}
	return originRHOAI
}

func isExternalOrigin(origin, kind string) bool {
	return origin == originExternal || strings.EqualFold(kind, "ExternalModel")
}

func lookupVendor(provider, model string) (VendorRate, bool) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	model = vendorModelKey(model)
	if model == "" {
		return VendorRate{}, false
	}
	var byName VendorRate
	foundName := false
	for _, v := range vendorRates {
		if vendorModelKey(v.Model) != model {
			continue
		}
		if provider != "" && provider == v.Provider {
			return v, true
		}
		if !foundName {
			byName = v
			foundName = true
		}
	}
	if provider == "" && foundName {
		return byName, true
	}
	return VendorRate{}, foundName && provider == ""
}

func vendorModelKey(model string) string {
	s := strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	return s
}

func blendedPerMillion(in, out float64) float64 {
	if in <= 0 && out <= 0 {
		return 0
	}
	if in <= 0 {
		return out
	}
	if out <= 0 {
		return in
	}
	// Limitador often has total tokens only; 3:1 input:output is a typical chat mix.
	return round4((3*in + out) / 4)
}
