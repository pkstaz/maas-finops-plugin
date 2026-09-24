package server

func mockOverview(window string, cat PricingCatalog) OverviewResponse {
	byModel := []UsageBreakdown{
		{Model: "qwen3-06b", TokensIn: 80_000, TokensOut: 48_400, Tokens: 128_400, Requests: 842, Limited: 12},
		{Model: "granite-8b", TokensIn: 40_000, TokensOut: 24_200, Tokens: 64_200, Requests: 310, Limited: 0},
	}
	bySub := []UsageBreakdown{
		{Subscription: "qwen-team", TokensIn: 70_000, TokensOut: 40_000, Tokens: 110_000, Requests: 500, Limited: 2},
		{Subscription: "qwen-authenticated", TokensIn: 40_000, TokensOut: 22_600, Tokens: 62_600, Requests: 500, Limited: 4},
		{Subscription: "qwen-free", TokensIn: 12_000, TokensOut: 8_000, Tokens: 20_000, Requests: 152, Limited: 6},
	}
	byUser := []UsageBreakdown{
		{User: "admin", TokensIn: 55_000, TokensOut: 35_000, Tokens: 90_000, Requests: 600, Limited: 4},
		{User: "workshop-user", TokensIn: 45_000, TokensOut: 27_600, Tokens: 72_600, Requests: 400, Limited: 6},
		{User: "finops-free", TokensIn: 18_000, TokensOut: 12_000, Tokens: 30_000, Requests: 152, Limited: 2},
	}
	price, _ := (&pricingStore{}).PriceFor(cat, "qwen3-06b")
	var totalIn, totalOut, totalTokens, totalReq, totalLim, totalCost float64
	top := []ModelRow{}
	for i := range byModel {
		p, priced := (&pricingStore{}).PriceFor(cat, byModel[i].Model)
		inP, outP := inputOutputPrices(cat, byModel[i].Model)
		byModel[i].Cost = costOfSplit(byModel[i].TokensIn, byModel[i].TokensOut, byModel[i].Tokens, p, inP, outP)
		totalIn += byModel[i].TokensIn
		totalOut += byModel[i].TokensOut
		totalTokens += byModel[i].Tokens
		totalReq += byModel[i].Requests
		totalLim += byModel[i].Limited
		totalCost += byModel[i].Cost
		top = append(top, ModelRow{
			Name: byModel[i].Model, DisplayName: byModel[i].Model, Namespace: "llm",
			Phase: "Ready", Kind: "MaaSModelRef",
			TokensIn: byModel[i].TokensIn, TokensOut: byModel[i].TokensOut, Tokens: byModel[i].Tokens,
			Requests: byModel[i].Requests, Limited: byModel[i].Limited,
			PricePerM: p, Cost: byModel[i].Cost, Priced: priced,
		})
	}
	for i := range bySub {
		bySub[i].Cost = costOf(bySub[i].Tokens, price)
	}
	for i := range byUser {
		byUser[i].Cost = costOf(byUser[i].Tokens, price)
	}
	return OverviewResponse{
		Range: window, Currency: cat.Currency,
		TotalTokensIn: totalIn, TotalTokensOut: totalOut, TotalTokens: totalTokens, TotalCost: totalCost,
		TotalRequests: totalReq, TotalLimited: totalLim, Source: "mock",
		TopModels: top, ByModel: byModel, BySubscription: bySub, ByUser: byUser,
	}
}

func mockModels(window string, cat PricingCatalog) ModelsResponse {
	items := []ModelRow{
		{Name: "qwen3-06b", Namespace: "llm", DisplayName: "Qwen/Qwen3-0.6B", Phase: "Ready", Kind: "LLMInferenceService", Origin: originRHOAI, TokensIn: 80000, TokensOut: 48400, Tokens: 128400, Requests: 842, Limited: 12},
		{Name: "gpt-4o", Namespace: "llm", DisplayName: "gpt-4o (Azure AI Foundry)", Phase: "Ready", Kind: "ExternalModel", Origin: originExternal, Provider: "azure-openai", TokensIn: 12000, TokensOut: 4000, Tokens: 16000, Requests: 40, Limited: 0},
	}
	for i := range items {
		p, inP, outP, priced := effectivePrices(cat, discoveredModel{
			Name: items[i].Name, Origin: items[i].Origin, Kind: items[i].Kind,
			Provider: items[i].Provider, TargetModel: items[i].Name,
		})
		items[i].PricePerM = p
		items[i].InputPerMillion = inP
		items[i].OutputPerMillion = outP
		items[i].Priced = priced
		items[i].Cost = costOfSplit(items[i].TokensIn, items[i].TokensOut, items[i].Tokens, p, inP, outP)
	}
	return ModelsResponse{Range: window, Currency: cat.Currency, Source: "mock", Items: items}
}

func mockSubscriptions(window string, cat PricingCatalog) SubscriptionsResponse {
	price := cat.DefaultPricePerMillion
	if mp, ok := cat.Models["qwen3-06b"]; ok {
		price = mp.PricePerMillion
	}
	items := []SubscriptionRow{
		{Name: "qwen-team", Namespace: "models-as-a-service", Priority: 50, Models: []string{"qwen3-06b"}, RateLimits: []string{"2000 / 30s", "20000 / 5m"}, TokensIn: 70000, TokensOut: 40000, Tokens: 110000, Requests: 500, Limited: 2},
		{Name: "qwen-authenticated", Namespace: "models-as-a-service", Priority: 10, Models: []string{"qwen3-06b"}, RateLimits: []string{"500 / 30s", "5000 / 5m"}, TokensIn: 40000, TokensOut: 22600, Tokens: 62600, Requests: 500, Limited: 4},
		{Name: "qwen-free", Namespace: "models-as-a-service", Priority: 1, Models: []string{"qwen3-06b"}, RateLimits: []string{"80 / 30s", "400 / 5m"}, TokensIn: 12000, TokensOut: 8000, Tokens: 20000, Requests: 152, Limited: 6},
	}
	for i := range items {
		items[i].Cost = costOf(items[i].Tokens, price)
	}
	return SubscriptionsResponse{Range: window, Currency: cat.Currency, Source: "mock", Items: items}
}

func mockAPIKeys(window string, cat PricingCatalog) ApiKeysResponse {
	price := cat.DefaultPricePerMillion
	items := []ApiKeyRow{
		{User: "admin", TokensIn: 55000, TokensOut: 35000, Tokens: 90000, Requests: 600, Limited: 4, Models: []string{"qwen3-06b", "granite-8b"}, Subscriptions: []string{"qwen-team", "qwen-authenticated"}},
		{User: "workshop-user", TokensIn: 45000, TokensOut: 27600, Tokens: 72600, Requests: 400, Limited: 6, Models: []string{"qwen3-06b"}, Subscriptions: []string{"qwen-authenticated"}},
		{User: "finops-free", TokensIn: 18000, TokensOut: 12000, Tokens: 30000, Requests: 152, Limited: 2, Models: []string{"qwen3-06b"}, Subscriptions: []string{"qwen-free"}},
	}
	for i := range items {
		items[i].Cost = costOf(items[i].Tokens, price)
	}
	return ApiKeysResponse{
		Range: window, Currency: cat.Currency, Source: "mock",
		Note:  "The Limitador user label is the MaaS API key owner (not the sk-oai-… id). Enable captureUser / TelemetryPolicy to see live data.",
		Items: items,
	}
}

func mockInventory() ClusterInventory {
	return ClusterInventory{
		Provider:  "azure",
		Platform:  "Azure",
		Region:    "eastus",
		Cloud:     "AzurePublicCloud",
		Source:    "mock",
		Supported: true,
		Nodes: []NodeInventory{
			{
				Name: "aro-gpu-eastus1", Role: "gpu", InstanceType: "Standard_NC8as_T4_v3",
				Region: "eastus", Zone: "eastus-1", GPUProduct: "Tesla T4", GPUFamily: "turing",
				GPUCount: 1, GPUMemoryMiB: 16384,
			},
		},
		Models: []ServingModel{
			{
				Name: "llama-31-instruct", Namespace: "llm",
				DisplayName: "Llama-3.1-8B-Instruct (T4 W4A16)",
				URI:         "hf://RedHatAI/Meta-Llama-3.1-8B-Instruct-quantized.w4a16",
				Quant:       "w4a16", ParamsB: 8, GPURequest: 1, MaxModelLen: 4096,
				NodeName: "aro-gpu-eastus1", Kind: "LLMInferenceService", Origin: originRHOAI,
			},
			{
				Name: "gpt-4o", Namespace: "llm", DisplayName: "gpt-4o (Azure AI Foundry)",
				Kind: "ExternalModel", Origin: originExternal, Provider: "azure-openai",
				Endpoint: "example.cognitiveservices.azure.com", TargetModel: "gpt-4o",
			},
			{
				Name: "glm-53-flash", Namespace: "llm", DisplayName: "glm-53-flash (TMM MaaS)",
				Kind: "ExternalModel", Origin: originExternal, Provider: "openai",
				Endpoint: "maas.example.dev", TargetModel: "glm-53-flash",
			},
		},
	}
}
