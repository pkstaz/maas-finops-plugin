package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIBMRegionKey(t *testing.T) {
	if ibmRegionKey("dal10") != "us-south" {
		t.Fatal(ibmRegionKey("dal10"))
	}
	if ibmRegionKey("us-south") != "us-south" {
		t.Fatal(ibmRegionKey("us-south"))
	}
}

func TestFetchIBMCatalog(t *testing.T) {
	liveCloudPrices = true
	t.Cleanup(func() { liveCloudPrices = true })
	ibmCache = &ibmPriceCache{values: map[string]CostSource{}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/gx3-24x120x1l40s/pricing"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"starting_price": map[string]string{"plan_id": "plan-1"},
			})
		case r.URL.Path == "/plan-1/pricing/deployment":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": []map[string]any{
					{
						"deployment_region": "au-syd",
						"metrics": []map[string]any{{
							"charge_unit_name": "INSTANCE_HOURS_MULTI_TENANT",
							"charge_unit":      "Instance-Hour",
							"amounts": []map[string]any{{
								"country": "USA", "currency": "USD",
								"prices": []map[string]any{{"price": 9.99}},
							}},
						}},
					},
					{
						"deployment_region": "us-south",
						"metrics": []map[string]any{{
							"charge_unit_name": "INSTANCE_HOURS_MULTI_TENANT",
							"charge_unit":      "Instance-Hour",
							"amounts": []map[string]any{{
								"country": "USA", "currency": "USD",
								"prices": []map[string]any{{"price": 3.94}},
							}},
						}},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	ibmCatalogAPI = srv.URL
	t.Cleanup(func() { ibmCatalogAPI = "https://globalcatalog.cloud.ibm.com/api/v1" })

	cost, ok := fetchIBMCatalog("gx3-24x120x1l40s", "dal10")
	if !ok || cost.HourlyUSD != 3.94 || cost.Kind != "ibm-catalog" {
		t.Fatalf("%+v ok=%v", cost, ok)
	}
	got := lookupMachineCost(HardwareConfig{}, "gx3-24x120x1l40s", "us-south", "ibmcloud")
	if got.Kind != "ibm-catalog" || got.HourlyUSD != 3.94 {
		t.Fatalf("lookup used %+v", got)
	}
}
