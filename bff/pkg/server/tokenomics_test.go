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
			// Limitador total only: in/out split unknown.
			TokensIn: 0, TokensOut: 0, Tokens: 642_200, Requests: 0,
			Cost: costOf(642_200, 0.15),
		},
		{
			Name: "gpt-4o", DisplayName: "gpt-4o (Azure AI Foundry)",
			Kind: "ExternalModel", Origin: originExternal, Provider: "azure-openai", TargetModel: "gpt-4o",
			TokensIn: 12_000, TokensOut: 4_000, Tokens: 16_000, Requests: 40,
			Cost: costOfSplit(12_000, 4_000, 16_000, 2.50, 2.50, 10),
		},
	}
	resp := tokenomicsFromRows("24h", cat, models, "")
	if resp.Tokens != 658_200 || resp.Requests != 40 {
		t.Fatalf("totals: %+v", resp)
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
	// Total-only row: blended 3:1 rates (10, 0.85, 3.4375) over 642200 tokens.
	wantBlended := []float64{6.422, 0.54587, 2.2075625}
	for i, want := range wantBlended {
		if diff := glm.Variants[i].Cost - want; diff < -2e-6 || diff > 2e-6 {
			t.Fatalf("glm variant %d cost %v want %v", i, glm.Variants[i].Cost, want)
		}
	}
	if glm.Variants[0].Model != "opus-4.8" || glm.Variants[1].Model != "gemini-flash" || glm.Variants[2].Model != "gpt-5" {
		t.Fatalf("variant models: %+v", glm.Variants)
	}
	gpt := resp.Items[1]
	// Split row: 12000/1e6*in + 4000/1e6*out.
	wantSplit := []float64{0.16, 0.0136, 0.055}
	for i, want := range wantSplit {
		if diff := gpt.Variants[i].Cost - want; diff < -2e-6 || diff > 2e-6 {
			t.Fatalf("gpt-4o variant %d cost %v want %v", i, gpt.Variants[i].Cost, want)
		}
	}
	// Provider totals sum both rows.
	wantTotals := []float64{6.582, 0.55947, 2.2625625}
	for i, want := range wantTotals {
		if diff := resp.Providers[i].Cost - want; diff < -2e-6 || diff > 2e-6 {
			t.Fatalf("provider %d total %v want %v", i, resp.Providers[i].Cost, want)
		}
	}
}
