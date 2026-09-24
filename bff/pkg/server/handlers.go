package server

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ok",
		Mock:      s.mock,
		ThanosURL: s.thanosURL,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, GetUser(r))
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	window := queryRange(r)
	cat := s.pricing.Get(r.Context())
	if s.mock {
		writeJSON(w, http.StatusOK, mockOverview(window, cat))
		return
	}
	models, modelErr := s.liveModels(r, window, cat)
	subs, subErr := s.liveSubscriptions(r, window, cat)
	keys, keyErr := s.liveAPIKeys(r, window, cat)

	var totalIn, totalOut, totalTokens, totalCost, totalReq, totalLim float64
	unpriced := 0
	byModel := make([]UsageBreakdown, 0, len(models))
	for _, m := range models {
		totalIn += m.TokensIn
		totalOut += m.TokensOut
		totalTokens += m.Tokens
		totalCost += m.Cost
		totalReq += m.Requests
		totalLim += m.Limited
		if !m.Priced {
			unpriced++
		}
		byModel = append(byModel, UsageBreakdown{
			Model: m.Name, TokensIn: m.TokensIn, TokensOut: m.TokensOut, Tokens: m.Tokens,
			Requests: m.Requests, Limited: m.Limited, Cost: m.Cost,
		})
	}
	bySub := make([]UsageBreakdown, 0, len(subs))
	var subReq, subLim float64
	for _, sub := range subs {
		subReq += sub.Requests
		subLim += sub.Limited
		bySub = append(bySub, UsageBreakdown{
			Subscription: sub.Name, TokensIn: sub.TokensIn, TokensOut: sub.TokensOut, Tokens: sub.Tokens,
			Requests: sub.Requests, Limited: sub.Limited, Cost: sub.Cost,
		})
	}
	if totalReq == 0 {
		totalReq = subReq
	}
	if totalLim == 0 {
		totalLim = subLim
	}
	byUser := make([]UsageBreakdown, 0, len(keys))
	for _, k := range keys {
		byUser = append(byUser, UsageBreakdown{
			User: k.User, TokensIn: k.TokensIn, TokensOut: k.TokensOut, Tokens: k.Tokens,
			Requests: k.Requests, Limited: k.Limited, Cost: k.Cost,
		})
	}
	sort.Slice(models, func(i, j int) bool { return models[i].Cost > models[j].Cost })
	top := models
	if len(top) > 5 {
		top = top[:5]
	}
	writeJSON(w, http.StatusOK, OverviewResponse{
		Range: window, Currency: cat.Currency,
		TotalTokensIn: totalIn, TotalTokensOut: totalOut, TotalTokens: totalTokens,
		TotalCost: totalCost, TotalRequests: totalReq, TotalLimited: totalLim, Unpriced: unpriced,
		Source: "cluster", MetricsError: firstPromErr(modelErr, subErr, keyErr),
		TopModels: top, ByModel: byModel, BySubscription: bySub, ByUser: byUser,
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	window := queryRange(r)
	cat := s.pricing.Get(r.Context())
	if s.mock {
		writeJSON(w, http.StatusOK, mockModels(window, cat))
		return
	}
	items, metricsErr := s.liveModels(r, window, cat)
	writeJSON(w, http.StatusOK, ModelsResponse{
		Range: window, Currency: cat.Currency, Source: "cluster", MetricsError: errString(metricsErr), Items: items,
	})
}

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	window := queryRange(r)
	cat := s.pricing.Get(r.Context())
	if s.mock {
		writeJSON(w, http.StatusOK, mockSubscriptions(window, cat))
		return
	}
	items, metricsErr := s.liveSubscriptions(r, window, cat)
	writeJSON(w, http.StatusOK, SubscriptionsResponse{
		Range: window, Currency: cat.Currency, Source: "cluster", MetricsError: errString(metricsErr), Items: items,
	})
}

func (s *Server) handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	window := queryRange(r)
	cat := s.pricing.Get(r.Context())
	if s.mock {
		writeJSON(w, http.StatusOK, mockAPIKeys(window, cat))
		return
	}
	items, metricsErr := s.liveAPIKeys(r, window, cat)
	writeJSON(w, http.StatusOK, ApiKeysResponse{
		Range: window, Currency: cat.Currency, Source: "cluster",
		Note:         "Limitador labels user with the MaaS API key owner (not the sk-oai-… secret). Without captureUser the series is empty. Input/output per user is not in Limitador; the vLLM split is per model.",
		MetricsError: errString(metricsErr),
		Items:        items,
	})
}

func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	cat := s.pricing.Get(r.Context())
	if u := r.URL.Query().Get("utilization"); u != "" {
		if f, err := strconv.ParseFloat(u, 64); err == nil {
			cat.Hardware.Utilization = f
		}
	}
	writeJSON(w, http.StatusOK, recommendPrices(s.clusterInventory(r), cat))
}

func (s *Server) handleRecommendApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user := GetUser(r)
	if !user.IsAdmin && !devMode {
		writeError(w, http.StatusForbidden, "only FinOps admins can apply prices")
		return
	}
	cat := s.pricing.Get(r.Context())
	rec := recommendPrices(s.clusterInventory(r), cat)
	applied := 0
	for _, m := range rec.Models {
		if !m.CanApply || m.PricePerMillion <= 0 {
			continue
		}
		prev := cat.Models[m.Name]
		prev.DisplayName = firstNonEmpty(m.DisplayName, prev.DisplayName)
		prev.PricePerMillion = m.PricePerMillion
		prev.InputPerMillion = m.InputPerMillion
		prev.OutputPerMillion = m.OutputPerMillion
		prev.Source = firstNonEmpty(m.Cost.Kind, "recommend")
		cat.Models[m.Name] = prev
		applied++
	}
	if applied == 0 {
		writeError(w, http.StatusUnprocessableEntity, "nothing to apply: missing GPU cost, vendor list, or throughput")
		return
	}
	if err := s.pricing.Put(r.Context(), cat); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"applied":   applied,
		"catalog":   s.pricing.Get(r.Context()),
		"recommend": rec,
	})
}

func (s *Server) clusterInventory(r *http.Request) ClusterInventory {
	if s.k8s != nil {
		return s.k8s.inventory(r.Context())
	}
	return mockInventory()
}

func (s *Server) handlePricing(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.pricing.Get(r.Context()))
	case http.MethodPut, http.MethodPost:
		user := GetUser(r)
		if !user.IsAdmin && !devMode {
			writeError(w, http.StatusForbidden, "only FinOps admins can change prices")
			return
		}
		var cat PricingCatalog
		if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.pricing.Put(r.Context(), cat); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, s.pricing.Get(r.Context()))
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) liveModels(r *http.Request, window string, cat PricingCatalog) ([]ModelRow, error) {
	ctx := r.Context()
	discovered, err := s.k8s.listModels(ctx)
	if err != nil {
		discovered = nil
	}
	tokensIn, errIn := s.prom.query(increaseBy(promptSelector(), "model_name,llm_isvc_name", window))
	tokensOut, errOut := s.prom.query(increaseBy(generationSelector(), "model_name,llm_isvc_name", window))
	tokens, errTok := s.prom.query(increaseBy(hitsSelector(), "model", window))
	reqs, errReq := s.prom.query(increaseBy(callsSelector(), "model", window))
	lim, errLim := s.prom.query(increaseBy(limitedSelector(), "model", window))

	byName := map[string]ModelRow{}
	for _, m := range discovered {
		price, inP, outP, priced := effectivePrices(cat, m)
		byName[m.Name] = ModelRow{
			Name: m.Name, Namespace: m.Namespace, DisplayName: m.DisplayName,
			Phase: m.Phase, Kind: m.Kind, Origin: m.Origin, Provider: m.Provider,
			Endpoint: m.Endpoint, TargetModel: m.TargetModel,
			PricePerM: price, InputPerMillion: inP, OutputPerMillion: outP, Priced: priced,
		}
	}
	mergeModelMetrics(byName, tokensIn, func(row *ModelRow, v float64) { row.TokensIn = v })
	mergeModelMetrics(byName, tokensOut, func(row *ModelRow, v float64) { row.TokensOut = v })
	mergeModelMetrics(byName, tokens, func(row *ModelRow, v float64) { row.Tokens = v })
	mergeModelMetrics(byName, reqs, func(row *ModelRow, v float64) { row.Requests = v })
	mergeModelMetrics(byName, lim, func(row *ModelRow, v float64) { row.Limited = v })

	out := make([]ModelRow, 0, len(byName))
	for _, row := range byName {
		row.TokensIn, row.TokensOut, row.Tokens = fillTokenTotals(row.TokensIn, row.TokensOut, row.Tokens)
		inP, outP := row.InputPerMillion, row.OutputPerMillion
		if inP == 0 && outP == 0 {
			inP, outP = inputOutputPrices(cat, row.Name)
		}
		row.Cost = costOfSplit(row.TokensIn, row.TokensOut, row.Tokens, row.PricePerM, inP, outP)
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tokens > out[j].Tokens })
	return out, firstError(errIn, errOut, errTok, errReq, errLim)
}

func (s *Server) liveSubscriptions(r *http.Request, window string, cat PricingCatalog) ([]SubscriptionRow, error) {
	ctx := r.Context()
	discovered, err := s.k8s.listSubscriptions(ctx)
	if err != nil {
		discovered = nil
	}
	tokens, errTok := s.prom.query(increaseBy(hitsSelector(), "subscription", window))
	reqs, errReq := s.prom.query(increaseBy(callsSelector(), "subscription", window))
	lim, errLim := s.prom.query(increaseBy(limitedSelector(), "subscription", window))

	byName := map[string]SubscriptionRow{}
	for _, sub := range discovered {
		byName[sub.Name] = SubscriptionRow{
			Name: sub.Name, Namespace: sub.Namespace, Priority: sub.Priority,
			Models: sub.Models, RateLimits: sub.RateLimits,
		}
	}
	applyNamed := func(vec metricVector, set func(*SubscriptionRow, float64)) {
		for key, v := range vec {
			_, name, _ := parseKey(key)
			if name == "" {
				name = key
			}
			if name == "" || name == "||" {
				continue
			}
			row := byName[name]
			if row.Name == "" {
				if v <= 0 {
					continue
				}
				row.Name = name
			}
			set(&row, math.Round(v))
			byName[name] = row
		}
	}
	applyNamed(tokens, func(row *SubscriptionRow, v float64) { row.Tokens = v })
	applyNamed(reqs, func(row *SubscriptionRow, v float64) { row.Requests = v })
	applyNamed(lim, func(row *SubscriptionRow, v float64) { row.Limited = v })

	out := make([]SubscriptionRow, 0, len(byName))
	for _, row := range byName {
		row.TokensIn, row.TokensOut, row.Tokens = fillTokenTotals(row.TokensIn, row.TokensOut, row.Tokens)
		price := cat.DefaultPricePerMillion
		inP, outP := 0.0, 0.0
		if len(row.Models) > 0 {
			if p, ok := s.pricing.PriceFor(cat, row.Models[0]); ok {
				price = p
			}
			inP, outP = inputOutputPrices(cat, row.Models[0])
		}
		row.Cost = costOfSplit(row.TokensIn, row.TokensOut, row.Tokens, price, inP, outP)
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tokens > out[j].Tokens })
	return out, firstError(errTok, errReq, errLim)
}

func (s *Server) liveAPIKeys(r *http.Request, window string, cat PricingCatalog) ([]ApiKeyRow, error) {
	tokens, errTok := s.prom.query(increaseBy(hitsSelector(), "user,subscription,model", window))
	reqs, errReq := s.prom.query(increaseBy(callsSelector(), "user,subscription,model", window))
	lim, errLim := s.prom.query(increaseBy(limitedSelector(), "user,subscription,model", window))

	type acc struct {
		row    ApiKeyRow
		models map[string]struct{}
		subs   map[string]struct{}
	}
	byUser := map[string]*acc{}
	ensure := func(user string) *acc {
		if user == "" {
			user = "(no user)"
		}
		a, ok := byUser[user]
		if !ok {
			a = &acc{
				row:    ApiKeyRow{User: user},
				models: map[string]struct{}{},
				subs:   map[string]struct{}{},
			}
			byUser[user] = a
		}
		return a
	}
	for key, v := range tokens {
		user, sub, model := parseKey(key)
		a := ensure(user)
		a.row.Tokens += math.Round(v)
		if model != "" {
			a.models[model] = struct{}{}
		}
		if sub != "" {
			a.subs[sub] = struct{}{}
		}
	}
	for key, v := range reqs {
		user, _, _ := parseKey(key)
		ensure(user).row.Requests += math.Round(v)
	}
	for key, v := range lim {
		user, _, _ := parseKey(key)
		ensure(user).row.Limited += math.Round(v)
	}

	out := make([]ApiKeyRow, 0, len(byUser))
	for _, a := range byUser {
		if a.row.Tokens <= 0 && a.row.Requests <= 0 && a.row.Limited <= 0 {
			continue
		}
		for m := range a.models {
			a.row.Models = append(a.row.Models, m)
		}
		for sub := range a.subs {
			a.row.Subscriptions = append(a.row.Subscriptions, sub)
		}
		sort.Strings(a.row.Models)
		sort.Strings(a.row.Subscriptions)
		a.row.TokensIn, a.row.TokensOut, a.row.Tokens = fillTokenTotals(a.row.TokensIn, a.row.TokensOut, a.row.Tokens)
		price := cat.DefaultPricePerMillion
		inP, outP := 0.0, 0.0
		if len(a.row.Models) > 0 {
			if p, ok := s.pricing.PriceFor(cat, a.row.Models[0]); ok {
				price = p
			}
			inP, outP = inputOutputPrices(cat, a.row.Models[0])
		}
		a.row.Cost = costOfSplit(a.row.TokensIn, a.row.TokensOut, a.row.Tokens, price, inP, outP)
		out = append(out, a.row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tokens > out[j].Tokens })
	return out, firstError(errTok, errReq, errLim)
}

func mergeModelMetrics(byName map[string]ModelRow, vec metricVector, set func(*ModelRow, float64)) {
	for key, v := range vec {
		_, _, model := parseKey(key)
		if model == "" || model == "||" {
			continue
		}
		short := model
		if i := strings.LastIndex(model, "/"); i >= 0 {
			short = model[i+1:]
		}
		row, ok := byName[short]
		if !ok {
			row, ok = byName[model]
		}
		if !ok {
			if v <= 0 {
				continue
			}
			row = ModelRow{Name: short, DisplayName: model, Kind: "metric"}
		}
		set(&row, math.Round(v))
		byName[row.Name] = row
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
