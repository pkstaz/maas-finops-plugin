package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSplitEC2Type(t *testing.T) {
	family, size, ok := splitEC2Type("g5.xlarge")
	if !ok || family != "g5" || size != "xlarge" {
		t.Fatalf("%s %s %v", family, size, ok)
	}
	family, size, ok = splitEC2Type("p4d.24xlarge")
	if !ok || family != "p4d" || size != "24xlarge" {
		t.Fatalf("%s %s %v", family, size, ok)
	}
	if _, _, ok := splitEC2Type("baremetal"); ok {
		t.Fatal("expected reject")
	}
}

func TestFetchAWSOnDemand(t *testing.T) {
	liveCloudPrices = true
	t.Cleanup(func() { liveCloudPrices = true })
	awsCache = &awsPriceCache{values: map[string]CostSource{}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/us-east-1/g5/xlarge.json" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"instanceType": "g5.xlarge", "price": 1.111})
	}))
	t.Cleanup(srv.Close)
	awsOnDemandBase = srv.URL
	t.Cleanup(func() { awsOnDemandBase = "https://www.ec2pricing.com" })

	cost, ok := fetchAWSOnDemand("g5.xlarge", "us-east-1")
	if !ok || cost.HourlyUSD != 1.111 || cost.Kind != "aws-ondemand" {
		t.Fatalf("%+v ok=%v", cost, ok)
	}
	got := lookupMachineCost(HardwareConfig{}, "g5.xlarge", "us-east-1", "aws")
	if got.Kind != "aws-ondemand" || got.HourlyUSD != 1.111 {
		t.Fatalf("lookup used %+v", got)
	}
}
