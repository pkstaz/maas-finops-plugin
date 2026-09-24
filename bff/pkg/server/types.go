package server

type UserInfo struct {
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
	IsAdmin  bool     `json:"is_admin"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Mock      bool   `json:"mock"`
	ThanosURL string `json:"thanosUrl,omitempty"`
}

type PricingCatalog struct {
	Currency               string                `json:"currency"`
	DefaultPricePerMillion float64               `json:"defaultPricePerMillion"`
	Models                 map[string]ModelPrice `json:"models"`
	Hardware               HardwareConfig        `json:"hardware"`
}

type HardwareConfig struct {
	Provider    string          `json:"provider"`
	Utilization float64         `json:"utilization"`
	Margin      float64         `json:"margin"`
	PriceType   string          `json:"priceType"`
	Machines    []ManualMachine `json:"machines"`
}

type ManualMachine struct {
	InstanceType string  `json:"instanceType"`
	Region       string  `json:"region,omitempty"`
	HourlyUSD    float64 `json:"hourlyUsd"`
	GPUProduct   string  `json:"gpuProduct,omitempty"`
	GPUCount     int     `json:"gpuCount,omitempty"`
	GPUMemoryGiB int     `json:"gpuMemoryGib,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

type ModelPrice struct {
	DisplayName      string  `json:"displayName,omitempty"`
	PricePerMillion  float64 `json:"pricePerMillion"`
	InputPerMillion  float64 `json:"inputPerMillion,omitempty"`
	OutputPerMillion float64 `json:"outputPerMillion,omitempty"`
	Source           string  `json:"source,omitempty"`
}

type ModelRow struct {
	Name             string  `json:"name"`
	Namespace        string  `json:"namespace"`
	DisplayName      string  `json:"displayName"`
	Phase            string  `json:"phase"`
	Kind             string  `json:"kind"`
	Origin           string  `json:"origin,omitempty"`
	Provider         string  `json:"provider,omitempty"`
	Endpoint         string  `json:"endpoint,omitempty"`
	TargetModel      string  `json:"targetModel,omitempty"`
	TokensIn         float64 `json:"tokensIn"`
	TokensOut        float64 `json:"tokensOut"`
	Tokens           float64 `json:"tokens"`
	Requests         float64 `json:"requests"`
	Limited          float64 `json:"limited"`
	PricePerM        float64 `json:"pricePerMillion"`
	InputPerMillion  float64 `json:"inputPerMillion,omitempty"`
	OutputPerMillion float64 `json:"outputPerMillion,omitempty"`
	Cost             float64 `json:"cost"`
	Priced           bool    `json:"priced"`
}

type SubscriptionRow struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Priority   int64    `json:"priority"`
	Models     []string `json:"models"`
	RateLimits []string `json:"rateLimits"`
	TokensIn   float64  `json:"tokensIn"`
	TokensOut  float64  `json:"tokensOut"`
	Tokens     float64  `json:"tokens"`
	Requests   float64  `json:"requests"`
	Limited    float64  `json:"limited"`
	Cost       float64  `json:"cost"`
}

type ApiKeyRow struct {
	User          string   `json:"user"`
	TokensIn      float64  `json:"tokensIn"`
	TokensOut     float64  `json:"tokensOut"`
	Tokens        float64  `json:"tokens"`
	Requests      float64  `json:"requests"`
	Limited       float64  `json:"limited"`
	Cost          float64  `json:"cost"`
	Models        []string `json:"models"`
	Subscriptions []string `json:"subscriptions"`
}

type UsageBreakdown struct {
	User         string  `json:"user,omitempty"`
	Subscription string  `json:"subscription,omitempty"`
	Model        string  `json:"model,omitempty"`
	TokensIn     float64 `json:"tokensIn"`
	TokensOut    float64 `json:"tokensOut"`
	Tokens       float64 `json:"tokens"`
	Requests     float64 `json:"requests"`
	Limited      float64 `json:"limited"`
	Cost         float64 `json:"cost"`
}

type OverviewResponse struct {
	Range          string           `json:"range"`
	Currency       string           `json:"currency"`
	TotalTokensIn  float64          `json:"totalTokensIn"`
	TotalTokensOut float64          `json:"totalTokensOut"`
	TotalTokens    float64          `json:"totalTokens"`
	TotalCost      float64          `json:"totalCost"`
	TotalRequests  float64          `json:"totalRequests"`
	TotalLimited   float64          `json:"totalLimited"`
	Unpriced       int              `json:"unpricedModels"`
	Source         string           `json:"source"`
	MetricsError   string           `json:"metricsError,omitempty"`
	TopModels      []ModelRow       `json:"topModels"`
	ByModel        []UsageBreakdown `json:"byModel"`
	BySubscription []UsageBreakdown `json:"bySubscription"`
	ByUser         []UsageBreakdown `json:"byUser"`
}

type ModelsResponse struct {
	Range        string     `json:"range"`
	Currency     string     `json:"currency"`
	Source       string     `json:"source"`
	MetricsError string     `json:"metricsError,omitempty"`
	Items        []ModelRow `json:"items"`
}

type SubscriptionsResponse struct {
	Range        string            `json:"range"`
	Currency     string            `json:"currency"`
	Source       string            `json:"source"`
	MetricsError string            `json:"metricsError,omitempty"`
	Items        []SubscriptionRow `json:"items"`
}

type ApiKeysResponse struct {
	Range        string      `json:"range"`
	Currency     string      `json:"currency"`
	Source       string      `json:"source"`
	Note         string      `json:"note"`
	MetricsError string      `json:"metricsError,omitempty"`
	Items        []ApiKeyRow `json:"items"`
}

type NodeInventory struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	InstanceType string `json:"instanceType"`
	Region       string `json:"region"`
	Zone         string `json:"zone"`
	GPUProduct   string `json:"gpuProduct"`
	GPUFamily    string `json:"gpuFamily"`
	GPUCount     int    `json:"gpuCount"`
	GPUMemoryMiB int    `json:"gpuMemoryMib"`
}

type ServingModel struct {
	Name        string  `json:"name"`
	Namespace   string  `json:"namespace"`
	DisplayName string  `json:"displayName"`
	URI         string  `json:"uri"`
	Quant       string  `json:"quant"`
	ParamsB     float64 `json:"paramsB"`
	GPURequest  float64 `json:"gpuRequest"`
	MaxModelLen int     `json:"maxModelLen"`
	NodeName    string  `json:"nodeName"`
	Kind        string  `json:"kind,omitempty"`
	Origin      string  `json:"origin,omitempty"`
	Provider    string  `json:"provider,omitempty"`
	Endpoint    string  `json:"endpoint,omitempty"`
	TargetModel string  `json:"targetModel,omitempty"`
}

type ClusterInventory struct {
	Provider  string          `json:"provider"`
	Platform  string          `json:"platform"`
	Region    string          `json:"region"`
	Cloud     string          `json:"cloud"`
	Source    string          `json:"source"`
	Supported bool            `json:"supported"`
	Message   string          `json:"message,omitempty"`
	Nodes     []NodeInventory `json:"nodes"`
	Models    []ServingModel  `json:"models"`
}

type CostSource struct {
	HourlyUSD float64 `json:"hourlyUsd"`
	Kind      string  `json:"kind"`
	SKU       string  `json:"sku"`
	Region    string  `json:"region"`
	Meter     string  `json:"meter,omitempty"`
	Note      string  `json:"note,omitempty"`
}

type ModelRecommendation struct {
	Name             string     `json:"name"`
	DisplayName      string     `json:"displayName"`
	Namespace        string     `json:"namespace"`
	Origin           string     `json:"origin,omitempty"`
	Kind             string     `json:"kind,omitempty"`
	Provider         string     `json:"provider,omitempty"`
	Endpoint         string     `json:"endpoint,omitempty"`
	InstanceType     string     `json:"instanceType"`
	GPUProduct       string     `json:"gpuProduct"`
	Quant            string     `json:"quant"`
	ParamsB          float64    `json:"paramsB"`
	TokensPerSecFull float64    `json:"tokensPerSecFull"`
	TokensPerSecUsed float64    `json:"tokensPerSecUsed"`
	TokensPerHour    float64    `json:"tokensPerHour"`
	HourlyUSD        float64    `json:"hourlyUsd"`
	PricePerMillion  float64    `json:"pricePerMillion"`
	InputPerMillion  float64    `json:"inputPerMillion,omitempty"`
	OutputPerMillion float64    `json:"outputPerMillion,omitempty"`
	Cost             CostSource `json:"cost"`
	Assumptions      []string   `json:"assumptions"`
	CanApply         bool       `json:"canApply"`
	Warning          string     `json:"warning,omitempty"`
}

type RecommendResponse struct {
	Inventory   ClusterInventory      `json:"inventory"`
	Utilization float64               `json:"utilization"`
	Margin      float64               `json:"margin"`
	Currency    string                `json:"currency"`
	Models      []ModelRecommendation `json:"models"`
	NeedsManual bool                  `json:"needsManual"`
	Notes       []string              `json:"notes"`
}
