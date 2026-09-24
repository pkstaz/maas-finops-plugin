package server

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Per-SKU Linux on-demand prices derived from the AWS Price List Bulk API
// (the official GetProducts API requires IAM; the regional offer JSON is ~480MiB).
var awsOnDemandBase = "https://www.ec2pricing.com"

type awsPriceCache struct {
	mu      sync.Mutex
	expires time.Time
	values  map[string]CostSource
}

var awsCache = &awsPriceCache{values: map[string]CostSource{}}

type awsOnDemandQuote struct {
	InstanceType string  `json:"instanceType"`
	Price        float64 `json:"price"`
}

func fetchAWSOnDemand(sku, region string) (CostSource, bool) {
	if !liveCloudPrices {
		return CostSource{}, false
	}
	if region == "" {
		region = "us-east-1"
	}
	family, size, ok := splitEC2Type(sku)
	if !ok {
		return CostSource{}, false
	}
	key := strings.ToLower(sku + "|" + region)
	awsCache.mu.Lock()
	if time.Now().Before(awsCache.expires) {
		if v, ok := awsCache.values[key]; ok {
			awsCache.mu.Unlock()
			return v, v.HourlyUSD > 0
		}
	}
	awsCache.mu.Unlock()

	u := fmt.Sprintf("%s/%s/%s/%s.json", strings.TrimRight(awsOnDemandBase, "/"), region, family, size)
	var quote awsOnDemandQuote
	if !getJSON(u, &quote) || quote.Price <= 0 {
		return CostSource{}, false
	}
	cs := CostSource{
		HourlyUSD: quote.Price,
		Kind:      "aws-ondemand",
		SKU:       sku,
		Region:    region,
		Note:      "AWS Price List Linux on-demand",
	}
	awsCache.mu.Lock()
	if awsCache.values == nil {
		awsCache.values = map[string]CostSource{}
	}
	awsCache.values[key] = cs
	awsCache.expires = time.Now().Add(6 * time.Hour)
	awsCache.mu.Unlock()
	return cs, true
}

func splitEC2Type(sku string) (family, size string, ok bool) {
	i := strings.IndexByte(sku, '.')
	if i <= 0 || i == len(sku)-1 {
		return "", "", false
	}
	return sku[:i], sku[i+1:], true
}
