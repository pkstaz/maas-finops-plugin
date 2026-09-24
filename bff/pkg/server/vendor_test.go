package server

import "testing"

func TestOriginFromKind(t *testing.T) {
	if originFromKind("ExternalModel") != originExternal {
		t.Fatal(originFromKind("ExternalModel"))
	}
	if originFromKind("LLMInferenceService") != originRHOAI {
		t.Fatal(originFromKind("LLMInferenceService"))
	}
	if !isExternalOrigin("external", "MaaSModelRef") {
		t.Fatal("origin should win")
	}
}

func TestLookupVendorGPT4o(t *testing.T) {
	v, ok := lookupVendor("azure-openai", "gpt-4o")
	if !ok || v.InputPerMillion != 2.50 || v.OutputPerMillion != 10 {
		t.Fatalf("%+v %v", v, ok)
	}
	if _, ok := lookupVendor("openai", "glm-53-flash"); ok {
		t.Fatal("unknown MaaS model should not match OpenAI gpt rates")
	}
	blend := blendedPerMillion(2.50, 10)
	if blend < 4.3 || blend > 4.4 {
		t.Fatalf("3:1 blend %v", blend)
	}
}

func TestRecommendSkipsGPUForExternal(t *testing.T) {
	inv := ClusterInventory{
		Provider:  "azure",
		Supported: true,
		Nodes: []NodeInventory{{
			Name: "gpu-1", InstanceType: "Standard_NC8as_T4_v3", GPUProduct: "Tesla T4", GPUCount: 1, Region: "eastus",
		}},
		Models: []ServingModel{
			{Name: "llama-31-instruct", DisplayName: "Llama", Origin: originRHOAI, Kind: "LLMInferenceService", Quant: "w4a16", ParamsB: 8, NodeName: "gpu-1"},
			{Name: "gpt-4o", DisplayName: "gpt-4o", Origin: originExternal, Kind: "ExternalModel", Provider: "azure-openai", Endpoint: "example.cognitiveservices.azure.com", TargetModel: "gpt-4o"},
			{Name: "glm-53-flash", DisplayName: "glm", Origin: originExternal, Kind: "ExternalModel", Provider: "openai", TargetModel: "glm-53-flash"},
		},
	}
	rec := recommendPrices(inv, PricingCatalog{
		Currency: "USD",
		Models:   map[string]ModelPrice{},
		Hardware: HardwareConfig{
			Utilization: 0.8,
			Machines:    []ManualMachine{{InstanceType: "Standard_NC8as_T4_v3", HourlyUSD: 0.752}},
		},
	})
	if len(rec.Models) != 3 {
		t.Fatalf("models %d", len(rec.Models))
	}
	byName := map[string]ModelRecommendation{}
	for _, m := range rec.Models {
		byName[m.Name] = m
	}
	llama := byName["llama-31-instruct"]
	if llama.Origin != originRHOAI || !llama.CanApply || llama.PricePerMillion < 3 {
		t.Fatalf("rhoai: %+v", llama)
	}
	gpt := byName["gpt-4o"]
	if gpt.Origin != originExternal || gpt.HourlyUSD != 0 || gpt.InstanceType != "" {
		t.Fatalf("external should not use GPU: %+v", gpt)
	}
	if !gpt.CanApply || gpt.InputPerMillion != 2.50 || gpt.OutputPerMillion != 10 {
		t.Fatalf("vendor gpt-4o: %+v", gpt)
	}
	glm := byName["glm-53-flash"]
	if glm.CanApply || glm.Cost.Kind != "missing" {
		t.Fatalf("unknown external needs manual token price: %+v", glm)
	}
}
