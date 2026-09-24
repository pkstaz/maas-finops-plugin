package server

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
)

type Server struct {
	distDir   string
	mock      bool
	k8s       *clusterClient
	prom      *promClient
	pricing   *pricingStore
	thanosURL string
}

func New(distDir string) (*Server, error) {
	mock := os.Getenv("MOCK_DATA") == "true"
	k8s, err := newClusterClient()
	if err != nil {
		if mock || os.Getenv("DEV_MODE") == "true" {
			k8s = nil
			mock = true
		} else {
			return nil, err
		}
	}
	prom := newPromClient()
	pricingClient := k8s
	if mock {
		pricingClient = nil
	}
	s := &Server{
		distDir:   distDir,
		mock:      mock,
		k8s:       k8s,
		prom:      prom,
		pricing:   newPricingStore(pricingClient),
		thanosURL: prom.baseURL,
	}
	return s, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)

	api := http.NewServeMux()
	api.HandleFunc("/api/auth/me", s.handleMe)
	api.HandleFunc("/api/overview", s.handleOverview)
	api.HandleFunc("/api/models", s.handleModels)
	api.HandleFunc("/api/subscriptions", s.handleSubscriptions)
	api.HandleFunc("/api/apikeys", s.handleAPIKeys)
	api.HandleFunc("/api/pricing", s.handlePricing)
	api.HandleFunc("/api/recommend", s.handleRecommend)
	api.HandleFunc("/api/recommend/apply", s.handleRecommendApply)
	api.HandleFunc("/api/simulator", s.handleSimulatorCatalog)
	api.HandleFunc("/api/simulator/quote", s.handleSimulatorQuote)

	mux.Handle("/api/", AuthMiddleware(api))

	if s.distDir != "" {
		fs := http.FileServer(http.Dir(s.distDir))
		mux.Handle("/", fs)
	}
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func queryRange(r *http.Request) string {
	q := r.URL.Query().Get("range")
	switch q {
	case "1h", "6h", "24h", "7d":
		return q
	default:
		return "24h"
	}
}

func costOf(tokens, pricePerMillion float64) float64 {
	if pricePerMillion <= 0 {
		return 0
	}
	v := (tokens / 1_000_000.0) * pricePerMillion
	return math.Round(v*1e6) / 1e6
}

func costOfSplit(tokensIn, tokensOut, tokensTotal, price, inPrice, outPrice float64) float64 {
	if inPrice > 0 || outPrice > 0 {
		in := inPrice
		out := outPrice
		if in <= 0 {
			in = price
		}
		if out <= 0 {
			out = price
		}
		return costOf(tokensIn, in) + costOf(tokensOut, out)
	}
	total := tokensTotal
	if total <= 0 {
		total = tokensIn + tokensOut
	}
	return costOf(total, price)
}

func fillTokenTotals(in, out, total float64) (float64, float64, float64) {
	in = math.Round(in)
	out = math.Round(out)
	total = math.Round(total)
	if in < 0 {
		in = 0
	}
	if out < 0 {
		out = 0
	}
	if total < 0 {
		total = 0
	}
	if in > 0 || out > 0 {
		total = in + out
	}
	return in, out, total
}

func inputOutputPrices(cat PricingCatalog, model string) (in, out float64) {
	if mp, ok := cat.Models[model]; ok {
		return mp.InputPerMillion, mp.OutputPerMillion
	}
	return 0, 0
}
