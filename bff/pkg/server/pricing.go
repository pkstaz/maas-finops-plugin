package server

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const pricingConfigMapName = "maas-finops-pricing"

type pricingStore struct {
	k8s *clusterClient
	mu  sync.Mutex
	mem PricingCatalog
}

func defaultCatalog() PricingCatalog {
	return PricingCatalog{
		Currency:               "USD",
		DefaultPricePerMillion: 0.15,
		Models:                 map[string]ModelPrice{},
		Hardware: HardwareConfig{
			Provider:    "auto",
			Utilization: 0.8,
			Margin:      0,
			PriceType:   "Consumption",
			Machines:    []ManualMachine{},
		},
	}
}

func newPricingStore(k8s *clusterClient) *pricingStore {
	cat := defaultCatalog()
	if raw := os.Getenv("DEFAULT_PRICING_JSON"); raw != "" {
		_ = json.Unmarshal([]byte(raw), &cat)
	}
	return &pricingStore{k8s: k8s, mem: cat}
}

func (p *pricingStore) Get(ctx context.Context) PricingCatalog {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.k8s != nil {
		cm, err := p.k8s.getConfigMap(ctx, pricingConfigMapName)
		if err == nil && cm.Data["pricing.json"] != "" {
			var cat PricingCatalog
			if json.Unmarshal([]byte(cm.Data["pricing.json"]), &cat) == nil {
				if cat.Models == nil {
					cat.Models = map[string]ModelPrice{}
				}
				if cat.Currency == "" {
					cat.Currency = "USD"
				}
				cat.Hardware = normalizeHardware(cat.Hardware)
				p.mem = cat
			}
		}
	}
	return cloneCatalog(p.mem)
}

func (p *pricingStore) Put(ctx context.Context, cat PricingCatalog) error {
	if cat.Models == nil {
		cat.Models = map[string]ModelPrice{}
	}
	if cat.Currency == "" {
		cat.Currency = "USD"
	}
	cat.Hardware = normalizeHardware(cat.Hardware)
	p.mu.Lock()
	p.mem = cat
	p.mu.Unlock()
	if p.k8s == nil {
		return nil
	}
	raw, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pricingConfigMapName,
			Namespace: p.k8s.ns,
			Labels: map[string]string{
				"app.kubernetes.io/name": "maas-finops-plugin",
			},
		},
		Data: map[string]string{"pricing.json": string(raw)},
	}
	return p.k8s.upsertConfigMap(ctx, cm)
}

func (p *pricingStore) PriceFor(cat PricingCatalog, model string) (float64, bool) {
	if mp, ok := cat.Models[model]; ok && mp.PricePerMillion > 0 {
		return mp.PricePerMillion, true
	}
	// MaaS ids sometimes look like publishers/llm/models/qwen3-06b
	for name, mp := range cat.Models {
		if name == model || suffixMatch(model, name) {
			if mp.PricePerMillion > 0 {
				return mp.PricePerMillion, true
			}
		}
	}
	return cat.DefaultPricePerMillion, cat.DefaultPricePerMillion > 0
}

func effectivePrices(cat PricingCatalog, m discoveredModel) (price, in, out float64, priced bool) {
	if mp, ok := cat.Models[m.Name]; ok {
		in, out = mp.InputPerMillion, mp.OutputPerMillion
		if mp.PricePerMillion > 0 || in > 0 || out > 0 {
			price = mp.PricePerMillion
			if price <= 0 {
				price = blendedPerMillion(in, out)
			}
			return price, in, out, true
		}
	}
	if isExternalOrigin(m.Origin, m.Kind) {
		if v, ok := lookupVendor(m.Provider, firstNonEmpty(m.TargetModel, m.Name)); ok {
			return blendedPerMillion(v.InputPerMillion, v.OutputPerMillion), v.InputPerMillion, v.OutputPerMillion, true
		}
		return 0, 0, 0, false
	}
	p, priced := (&pricingStore{}).PriceFor(cat, m.Name)
	return p, 0, 0, priced
}

func suffixMatch(full, short string) bool {
	if full == short {
		return true
	}
	if len(full) > len(short) && (full[len(full)-len(short)-1] == '/' || full[len(full)-len(short)-1] == '-') {
		return full[len(full)-len(short):] == short
	}
	return false
}

func cloneCatalog(in PricingCatalog) PricingCatalog {
	out := in
	out.Models = map[string]ModelPrice{}
	for k, v := range in.Models {
		out.Models[k] = v
	}
	out.Hardware.Machines = append([]ManualMachine{}, in.Hardware.Machines...)
	return out
}
