package server

import "testing"

func TestSimulateQuoteT4Llama8B(t *testing.T) {
	q, err := simulateQuote("aws", "g4dn.2xlarge", "llama-3.1-8b-instruct-w4a16", 0.8, 0, PricingCatalog{})
	if err != nil {
		t.Fatal(err)
	}
	if q.TokensPerSecFull < 79 || q.TokensPerSecFull > 81 {
		t.Fatalf("expected ~80 tok/s at 100%%, got %v", q.TokensPerSecFull)
	}
	if q.PricePerMillion < 3.2 || q.PricePerMillion > 3.4 {
		t.Fatalf("expected ~$3.26 / 1M, got %v", q.PricePerMillion)
	}
	if !q.Fits {
		t.Fatal("8B W4A16 should fit on 16 GiB T4")
	}
}

func TestSimulateQuoteBaremetalNeedsHourly(t *testing.T) {
	if _, err := simulateQuote("baremetal", "1x Tesla T4", "llama-3.1-8b-instruct-w4a16", 0.8, 0, PricingCatalog{}); err == nil {
		t.Fatal("bare metal without USD/hour should fail")
	}
	q, err := simulateQuote("baremetal", "1x Tesla T4", "llama-3.1-8b-instruct-w4a16", 0.8, 0.752, PricingCatalog{})
	if err != nil {
		t.Fatal(err)
	}
	if q.PricePerMillion < 3.2 || q.PricePerMillion > 3.4 {
		t.Fatalf("bare metal quote %v", q.PricePerMillion)
	}
}

func TestFindCatalogModel(t *testing.T) {
	if _, ok := findCatalogModel("granite-3.3-8b-instruct"); !ok {
		t.Fatal("missing granite")
	}
	if _, ok := findSKU("azure", "Standard_NC8as_T4_v3"); !ok {
		t.Fatal("missing azure T4 SKU")
	}
}
