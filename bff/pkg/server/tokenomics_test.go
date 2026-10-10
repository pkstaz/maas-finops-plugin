package server

import (
	"testing"
)

func TestMatchVendorRatesGPT4o(t *testing.T) {
	rates := matchVendorRates("gpt-4o", "gpt-4o (Azure AI Foundry)", "gpt-4o")
	providers := map[string]VendorRate{}
	for _, v := range rates {
		providers[v.Provider] = v
	}
	if len(rates) != 4 {
		t.Fatalf("gpt-4o should match azure-foundry, azure-openai, openai, copilot; got %d: %+v", len(rates), rates)
	}
	if providers["azure-foundry"].InputPerMillion != 2.50 || providers["azure-foundry"].OutputPerMillion != 10 {
		t.Fatalf("azure-foundry gpt-4o: %+v", providers["azure-foundry"])
	}
	if providers["openai"].Note != "OpenAI API list" {
		t.Fatalf("openai gpt-4o: %+v", providers["openai"])
	}
	if providers["copilot"].PerRequest != 0 || providers["copilot"].Note == "" {
		t.Fatalf("copilot gpt-4o is a 0x base model: %+v", providers["copilot"])
	}
	if _, ok := providers["bedrock"]; ok {
		t.Fatal("bedrock has no gpt-4o variant")
	}
	if _, ok := providers["anthropic"]; ok {
		t.Fatal("anthropic has no gpt-4o variant")
	}
}

func TestMatchVendorRatesLongestKeyWins(t *testing.T) {
	rates := matchVendorRates("gpt-4o-mini", "gpt-4o-mini", "gpt-4o-mini")
	providers := map[string]VendorRate{}
	for _, v := range rates {
		providers[v.Provider] = v
	}
	if len(rates) != 3 {
		t.Fatalf("gpt-4o-mini should not match copilot; got %d: %+v", len(rates), rates)
	}
	if providers["openai"].Model != "gpt-4o-mini" {
		t.Fatalf("gpt-4o-mini must win over gpt-4o: %+v", providers["openai"])
	}
}

func TestMatchVendorRatesLlamaByDisplayName(t *testing.T) {
	rates := matchVendorRates("llama-31-instruct", "Llama-3.1-8B-Instruct (T4 W4A16)", "")
	providers := map[string]VendorRate{}
	for _, v := range rates {
		providers[v.Provider] = v
	}
	if _, ok := providers["azure-foundry"]; !ok {
		t.Fatalf("azure-foundry llama 8b should match by display name: %+v", rates)
	}
	if providers["azure-foundry"].InputPerMillion != 0.30 {
		t.Fatalf("azure-foundry llama 8b: %+v", providers["azure-foundry"])
	}
	if _, ok := providers["bedrock"]; !ok {
		t.Fatalf("bedrock llama 8b should match by display name: %+v", rates)
	}
	if providers["bedrock"].InputPerMillion != 0.22 {
		t.Fatalf("bedrock llama 8b: %+v", providers["bedrock"])
	}
}

func TestMatchVendorRatesUnknownModel(t *testing.T) {
	if rates := matchVendorRates("glm-53-flash", "glm-53-flash (TMM MaaS)", "glm-53-flash"); len(rates) != 0 {
		t.Fatalf("unknown MaaS model should not match any vendor: %+v", rates)
	}
	if rates := matchVendorRates("qwen3-06b", "Qwen/Qwen3-0.6B", "qwen3-06b"); len(rates) != 0 {
		t.Fatalf("open-weight model without public variant should not match: %+v", rates)
	}
}

func TestTokenomicsFromRows(t *testing.T) {
	cat := defaultCatalog()
	models := []ModelRow{
		{
			Name: "glm-53-flash", DisplayName: "glm-53-flash (TMM MaaS)",
			Kind: "ExternalModel", Origin: originExternal, Provider: "openai", TargetModel: "glm-53-flash",
			TokensIn: 80_000, TokensOut: 48_400, Tokens: 128_400, Requests: 842,
			Cost: costOfSplit(80_000, 48_400, 128_400, 0.15, 0, 0),
		},
		{
			Name: "gpt-4o", DisplayName: "gpt-4o (Azure AI Foundry)",
			Kind: "ExternalModel", Origin: originExternal, Provider: "azure-openai", TargetModel: "gpt-4o",
			TokensIn: 12_000, TokensOut: 4_000, Tokens: 16_000, Requests: 40,
			Cost: costOfSplit(12_000, 4_000, 16_000, 2.50, 2.50, 10),
		},
	}
	resp := tokenomicsFromRows("24h", cat, models, "")
	if resp.Tokens != 144_400 || resp.Requests != 882 {
		t.Fatalf("totals: %+v", resp)
	}
	wantMaas := costOfSplit(80_000, 48_400, 128_400, 0.15, 0, 0) + costOfSplit(12_000, 4_000, 16_000, 2.50, 2.50, 10)
	if diff := resp.MaasCost - wantMaas; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("maas cost %v want %v", resp.MaasCost, wantMaas)
	}
	if len(resp.Providers) != 3 {
		t.Fatalf("providers: %+v", resp.Providers)
	}
	wantOrder := []string{"anthropic", "google", "openai"}
	for i, id := range wantOrder {
		if resp.Providers[i].ID != id || resp.Providers[i].Mode != "tokens" {
			t.Fatalf("provider %d: %+v", i, resp.Providers[i])
		}
	}
	glm := resp.Items[0]
	if len(glm.Variants) != 3 {
		t.Fatalf("glm variants: %+v", glm.Variants)
	}
	// 80000/1e6*5 + 48400/1e6*25 = 1.61 at Opus 4.8.
	if diff := glm.Variants[0].Cost - 1.61; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("anthropic glm cost %v", glm.Variants[0].Cost)
	}
	// 80000/1e6*0.30 + 48400/1e6*2.50 = 0.145 at Gemini Flash.
	if diff := glm.Variants[1].Cost - 0.145; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("google glm cost %v", glm.Variants[1].Cost)
	}
	// 80000/1e6*1.25 + 48400/1e6*10 = 0.584 at GPT-5.
	if diff := glm.Variants[2].Cost - 0.584; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("openai glm cost %v", glm.Variants[2].Cost)
	}
	if glm.Variants[0].Model != "opus-4.8" || glm.Variants[1].Model != "gemini-flash" || glm.Variants[2].Model != "gpt-5" {
		t.Fatalf("variant models: %+v", glm.Variants)
	}
	// Provider totals sum both rows.
	if diff := resp.Providers[0].Cost - 1.77; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("anthropic total %v", resp.Providers[0].Cost)
	}
	if diff := resp.Providers[1].Cost - 0.1586; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("google total %v", resp.Providers[1].Cost)
	}
	if diff := resp.Providers[2].Cost - 0.639; diff < -1e-9 || diff > 1e-9 {
		t.Fatalf("openai total %v", resp.Providers[2].Cost)
	}
}
