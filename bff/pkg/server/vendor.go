package server

import (
	"sort"
	"strings"
)

const (
	originRHOAI    = "rhoai"
	originExternal = "external"
)

type VendorRate struct {
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	InputPerMillion  float64 `json:"inputPerMillion,omitempty"`
	OutputPerMillion float64 `json:"outputPerMillion,omitempty"`
	PerRequest       float64 `json:"perRequest,omitempty"`
	Note             string  `json:"note"`
}

// Public list USD / 1M tokens. Confirm in the vendor account; EA/PTU overrides go in the catalog.
// Copilot is per-request: premium requests cost $0.04 for a 1x model, so PerRequest = 0.04 x multiplier.
var vendorRates = []VendorRate{
	{"azure-foundry", "gpt-4o", 2.50, 10.00, 0, "Azure AI Foundry list"},
	{"azure-foundry", "gpt-4o-mini", 0.15, 0.60, 0, "Azure AI Foundry list"},
	{"azure-foundry", "gpt-4.1", 2.00, 8.00, 0, "Azure AI Foundry list"},
	{"azure-foundry", "gpt-4.1-mini", 0.40, 1.60, 0, "Azure AI Foundry list"},
	{"azure-foundry", "gpt-4.1-nano", 0.10, 0.40, 0, "Azure AI Foundry list"},
	{"azure-foundry", "Meta-Llama-3.1-8B-Instruct", 0.30, 0.30, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "Meta-Llama-3.1-70B-Instruct", 0.60, 0.60, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "Meta-Llama-3.3-70B-Instruct", 0.60, 0.60, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "Meta-Llama-3.1-405B-Instruct", 3.50, 3.50, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "DeepSeek-V3", 1.14, 4.40, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "DeepSeek-R1", 1.34, 5.35, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "Phi-4", 0.07, 0.14, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "Mistral-Large-2411", 2.00, 6.00, 0, "Azure AI Foundry serverless"},
	{"azure-foundry", "grok-3", 3.00, 15.00, 0, "Azure AI Foundry list"},
	{"azure-openai", "gpt-4o", 2.50, 10.00, 0, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4o-mini", 0.15, 0.60, 0, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1", 2.00, 8.00, 0, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1-mini", 0.40, 1.60, 0, "Azure OpenAI standard list"},
	{"azure-openai", "gpt-4.1-nano", 0.10, 0.40, 0, "Azure OpenAI standard list"},
	{"bedrock", "anthropic.claude-3-5-sonnet", 3.00, 15.00, 0, "AWS Bedrock on-demand"},
	{"bedrock", "anthropic.claude-3-5-haiku", 0.80, 4.00, 0, "AWS Bedrock on-demand"},
	{"bedrock", "anthropic.claude-3-haiku", 0.25, 1.25, 0, "AWS Bedrock on-demand"},
	{"bedrock", "anthropic.claude-sonnet-4", 3.00, 15.00, 0, "AWS Bedrock on-demand"},
	{"bedrock", "anthropic.claude-opus-4", 15.00, 75.00, 0, "AWS Bedrock on-demand"},
	{"bedrock", "meta.llama3-1-8b-instruct", 0.22, 0.22, 0, "AWS Bedrock on-demand"},
	{"bedrock", "meta.llama3-1-70b-instruct", 0.72, 0.72, 0, "AWS Bedrock on-demand"},
	{"bedrock", "meta.llama3-3-70b-instruct", 0.72, 0.72, 0, "AWS Bedrock on-demand"},
	{"bedrock", "meta.llama3-2-1b-instruct", 0.13, 0.13, 0, "AWS Bedrock on-demand"},
	{"bedrock", "meta.llama3-2-3b-instruct", 0.15, 0.20, 0, "AWS Bedrock on-demand"},
	{"bedrock", "amazon.nova-micro", 0.035, 0.14, 0, "AWS Bedrock on-demand"},
	{"bedrock", "amazon.nova-lite", 0.06, 0.24, 0, "AWS Bedrock on-demand"},
	{"bedrock", "amazon.nova-pro", 0.80, 3.20, 0, "AWS Bedrock on-demand"},
	{"bedrock", "deepseek.r1", 1.35, 5.40, 0, "AWS Bedrock on-demand"},
	{"openai", "gpt-4o", 2.50, 10.00, 0, "OpenAI API list"},
	{"openai", "gpt-4o-mini", 0.15, 0.60, 0, "OpenAI API list"},
	{"openai", "gpt-4.1", 2.00, 8.00, 0, "OpenAI API list"},
	{"openai", "gpt-4.1-mini", 0.40, 1.60, 0, "OpenAI API list"},
	{"openai", "gpt-4.1-nano", 0.10, 0.40, 0, "OpenAI API list"},
	{"openai", "o3", 2.00, 8.00, 0, "OpenAI API list"},
	{"openai", "o3-mini", 1.10, 4.40, 0, "OpenAI API list"},
	{"openai", "o4-mini", 1.10, 4.40, 0, "OpenAI API list"},
	{"anthropic", "claude-3-5-sonnet", 3.00, 15.00, 0, "Anthropic API list"},
	{"anthropic", "claude-3-5-haiku", 0.80, 4.00, 0, "Anthropic API list"},
	{"anthropic", "claude-3-opus", 15.00, 75.00, 0, "Anthropic API list"},
	{"anthropic", "claude-sonnet-4", 3.00, 15.00, 0, "Anthropic API list"},
	{"anthropic", "claude-opus-4", 15.00, 75.00, 0, "Anthropic API list"},
	{"copilot", "gpt-4.1", 0, 0, 0, "Copilot base model, included in plan"},
	{"copilot", "gpt-4o", 0, 0, 0, "Copilot base model, included in plan"},
	{"copilot", "claude-3-5-sonnet", 0, 0, 0.04, "Copilot 1x premium request"},
	{"copilot", "claude-3-7-sonnet", 0, 0, 0.04, "Copilot 1x premium request"},
	{"copilot", "claude-sonnet-4", 0, 0, 0.04, "Copilot 1x premium request"},
	{"copilot", "o1", 0, 0, 0.04, "Copilot 1x premium request"},
	{"copilot", "o3", 0, 0, 0.04, "Copilot 1x premium request"},
	{"copilot", "claude-opus-4", 0, 0, 0.40, "Copilot 10x premium request"},
	{"copilot", "gpt-4.5", 0, 0, 2.00, "Copilot 50x premium request"},
	{"copilot", "gemini-2.0-flash", 0, 0, 0.01, "Copilot 0.25x premium request"},
}

// tokenomicsProviders is the fixed display order for the comparison table.
var tokenomicsProviders = []struct {
	ID    string
	Label string
	Mode  string
}{
	{"azure-foundry", "Azure Foundry", "tokens"},
	{"azure-openai", "Azure OpenAI", "tokens"},
	{"bedrock", "AWS Bedrock", "tokens"},
	{"openai", "OpenAI", "tokens"},
	{"anthropic", "Anthropic Claude", "tokens"},
	{"copilot", "GitHub Copilot", "requests"},
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

// tokenomicsModelKey normalizes a model name for matching across providers:
// lowercase, strip any path prefix, keep only alphanumeric characters so
// "Qwen/Qwen3-0.6B" and "qwen3-06b" collapse to the same key.
func tokenomicsModelKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// matchVendorRates returns, for every provider, the most specific public
// variant matching the model (name, display name, or target model). Matching
// is by normalized key containment so quantized or dated names still resolve;
// publisher prefixes (Meta-, amazon., anthropic.) are stripped from vendor
// keys. Only variants matching the most specific key overall are kept, so
// gpt-4o-mini does not fall back to gpt-4o rates.
func matchVendorRates(name, displayName, targetModel string) []VendorRate {
	candidates := []string{name, displayName, targetModel}
	keys := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if k := tokenomicsModelKey(c); k != "" {
			keys = append(keys, k)
		}
	}
	type match struct {
		rate   VendorRate
		keyLen int
	}
	matches := []match{}
	for _, v := range vendorRates {
		for _, vk := range vendorKeys(v.Model) {
			if vk == "" {
				continue
			}
			found := false
			for _, k := range keys {
				if k == vk || strings.Contains(k, vk) {
					found = true
					break
				}
			}
			if found {
				matches = append(matches, match{rate: v, keyLen: len(vk)})
				break
			}
		}
	}
	if len(matches) == 0 {
		return nil
	}
	maxLen := 0
	for _, m := range matches {
		if m.keyLen > maxLen {
			maxLen = m.keyLen
		}
	}
	best := map[string]VendorRate{}
	for _, m := range matches {
		if m.keyLen != maxLen {
			continue
		}
		if prev, ok := best[m.rate.Provider]; !ok || m.keyLen > len(tokenomicsModelKey(prev.Model)) {
			best[m.rate.Provider] = m.rate
		}
	}
	out := make([]VendorRate, 0, len(best))
	for _, v := range best {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out
}

// vendorKeys returns the normalized keys to match a vendor model against:
// the full key plus the key with the publisher prefix removed.
func vendorKeys(model string) []string {
	vk := tokenomicsModelKey(model)
	keys := []string{vk}
	for _, p := range []string{"meta", "amazon", "anthropic", "mistralai", "cohere", "ai21", "nvidia"} {
		if strings.HasPrefix(vk, p) && len(vk) > len(p) {
			keys = append(keys, vk[len(p):])
			break
		}
	}
	return keys
}
