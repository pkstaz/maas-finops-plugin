package server

import (
	"net/url"
	"strings"
	"sync"
	"time"
)

// IBM Global Catalog (public GETs, no IAM token required for list pricing).
var ibmCatalogAPI = "https://globalcatalog.cloud.ibm.com/api/v1"

type ibmPriceCache struct {
	mu      sync.Mutex
	expires time.Time
	values  map[string]CostSource
}

var ibmCache = &ibmPriceCache{values: map[string]CostSource{}}

type ibmPricingHint struct {
	StartingPrice struct {
		PlanID string `json:"plan_id"`
	} `json:"starting_price"`
}

type ibmPlanList struct {
	Resources []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"resources"`
}

type ibmDeploymentList struct {
	Resources []ibmDeployment `json:"resources"`
}

type ibmDeployment struct {
	DeploymentID       string      `json:"deployment_id"`
	DeploymentLocation string      `json:"deployment_location"`
	DeploymentRegion   string      `json:"deployment_region"`
	Metrics            []ibmMetric `json:"metrics"`
}

type ibmMetric struct {
	ChargeUnit     string `json:"charge_unit"`
	ChargeUnitName string `json:"charge_unit_name"`
	Amounts        []struct {
		Country  string `json:"country"`
		Currency string `json:"currency"`
		Prices   []struct {
			Price float64 `json:"price"`
		} `json:"prices"`
	} `json:"amounts"`
}

func fetchIBMCatalog(sku, region string) (CostSource, bool) {
	if !liveCloudPrices {
		return CostSource{}, false
	}
	if region == "" {
		region = "us-south"
	}
	key := strings.ToLower(sku + "|" + region)
	ibmCache.mu.Lock()
	if time.Now().Before(ibmCache.expires) {
		if v, ok := ibmCache.values[key]; ok {
			ibmCache.mu.Unlock()
			return v, v.HourlyUSD > 0
		}
	}
	ibmCache.mu.Unlock()

	planID := ibmPlanID(sku)
	if planID == "" {
		return CostSource{}, false
	}
	var list ibmDeploymentList
	if !getJSON(ibmCatalogAPI+"/"+planID+"/pricing/deployment", &list) {
		return CostSource{}, false
	}
	dep, ok := pickIBMDeployment(list.Resources, region)
	if !ok {
		return CostSource{}, false
	}
	hourly := ibmHourlyUSD(dep.Metrics)
	if hourly <= 0 {
		return CostSource{}, false
	}
	cs := CostSource{
		HourlyUSD: hourly,
		Kind:      "ibm-catalog",
		SKU:       sku,
		Region:    firstNonEmpty(dep.DeploymentRegion, region),
		Meter:     "INSTANCE_HOURS",
		Note:      "IBM Global Catalog VPC instance-hour",
	}
	ibmCache.mu.Lock()
	if ibmCache.values == nil {
		ibmCache.values = map[string]CostSource{}
	}
	ibmCache.values[key] = cs
	ibmCache.expires = time.Now().Add(6 * time.Hour)
	ibmCache.mu.Unlock()
	return cs, true
}

func ibmPlanID(sku string) string {
	var hint ibmPricingHint
	if getJSON(ibmCatalogAPI+"/"+sku+"/pricing", &hint) && hint.StartingPrice.PlanID != "" {
		return hint.StartingPrice.PlanID
	}
	var plans ibmPlanList
	if !getJSON(ibmCatalogAPI+"/is.instance/plan?q="+url.QueryEscape(sku), &plans) {
		return ""
	}
	for _, p := range plans.Resources {
		if strings.EqualFold(p.Name, sku) || strings.EqualFold(p.ID, sku) {
			return p.ID
		}
	}
	if len(plans.Resources) == 1 {
		return plans.Resources[0].ID
	}
	return ""
}

func pickIBMDeployment(deps []ibmDeployment, region string) (ibmDeployment, bool) {
	want := ibmRegionKey(region)
	var fallback ibmDeployment
	var hasFallback bool
	for _, d := range deps {
		loc := ibmRegionKey(firstNonEmpty(d.DeploymentRegion, d.DeploymentLocation))
		if loc == want {
			return d, true
		}
		if !hasFallback && ibmHourlyUSD(d.Metrics) > 0 {
			fallback = d
			hasFallback = true
		}
	}
	return fallback, hasFallback
}

func ibmHourlyUSD(metrics []ibmMetric) float64 {
	var instance, total float64
	for _, m := range metrics {
		p := ibmUSDAmount(m)
		if p <= 0 {
			continue
		}
		unit := strings.ToUpper(m.ChargeUnitName + " " + m.ChargeUnit)
		if strings.Contains(unit, "INSTANCE") {
			instance = p
			break
		}
		total += p
	}
	if instance > 0 {
		return instance
	}
	return total
}

func ibmUSDAmount(m ibmMetric) float64 {
	var generic float64
	for _, a := range m.Amounts {
		if !strings.EqualFold(a.Currency, "USD") || len(a.Prices) == 0 || a.Prices[0].Price <= 0 {
			continue
		}
		country := strings.ToUpper(a.Country)
		if country == "USA" || country == "USD" || country == "US" {
			return a.Prices[0].Price
		}
		if generic == 0 {
			generic = a.Prices[0].Price
		}
	}
	return generic
}

func ibmRegionKey(region string) string {
	r := strings.ToLower(strings.TrimSpace(region))
	switch {
	case r == "us-south" || strings.Contains(r, "dal"):
		return "us-south"
	case r == "us-east" || strings.Contains(r, "wdc"):
		return "us-east"
	case r == "eu-de" || strings.Contains(r, "fra"):
		return "eu-de"
	case r == "eu-gb" || strings.Contains(r, "lon"):
		return "eu-gb"
	case r == "eu-es" || strings.Contains(r, "mad"):
		return "eu-es"
	case r == "jp-tok" || strings.Contains(r, "tok"):
		return "jp-tok"
	case r == "au-syd" || strings.Contains(r, "syd"):
		return "au-syd"
	case r == "br-sao" || strings.Contains(r, "sao"):
		return "br-sao"
	case r == "ca-tor" || strings.Contains(r, "tor"):
		return "ca-tor"
	default:
		return r
	}
}
