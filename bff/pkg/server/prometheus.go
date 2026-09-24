package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultThanosURL = "http://thanos-querier-data-science-thanos-querier.redhat-ods-monitoring.svc:10902"

type promClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type promResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
	Error string `json:"error"`
}

type metricVector map[string]float64

func newPromClient() *promClient {
	base := os.Getenv("THANOS_URL")
	if base == "" {
		base = defaultThanosURL
	}
	base = normalizeThanosURL(base)
	token := ""
	if b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		token = strings.TrimSpace(string(b))
	}
	return &promClient{
		baseURL: strings.TrimRight(base, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		},
	}
}

// RHOAI Thanos querier exposes HTTP on :10902 (port name "http"), not TLS.
func normalizeThanosURL(base string) string {
	if strings.Contains(base, ":10902") && strings.HasPrefix(base, "https://") {
		return "http://" + strings.TrimPrefix(base, "https://")
	}
	return base
}

func (p *promClient) query(q string) (metricVector, error) {
	if p == nil || p.baseURL == "" {
		return nil, fmt.Errorf("prometheus not configured")
	}
	u, err := url.Parse(p.baseURL + "/api/v1/query")
	if err != nil {
		return nil, err
	}
	qs := u.Query()
	qs.Set("query", q)
	u.RawQuery = qs.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("prometheus query error: %v q=%s", err, q)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		err := fmt.Errorf("prometheus %s: %s", resp.Status, string(body))
		log.Printf("%v q=%s", err, q)
		return nil, err
	}
	var parsed promResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Status != "success" {
		err := fmt.Errorf("prometheus: %s", firstNonEmpty(parsed.Error, parsed.Status))
		log.Printf("%v q=%s", err, q)
		return nil, err
	}
	out := metricVector{}
	for _, r := range parsed.Data.Result {
		if len(r.Value) < 2 {
			continue
		}
		s, _ := r.Value[1].(string)
		v, _ := strconv.ParseFloat(s, 64)
		if v < 0 {
			v = 0
		}
		key := joinLabels(r.Metric)
		out[key] += v
	}
	return out, nil
}

func joinLabels(m map[string]string) string {
	parts := []string{
		m["user"],
		m["subscription"],
		firstNonEmpty(m["model"], m["model_name"], m["llm_isvc_name"]),
	}
	return strings.Join(parts, "|")
}

func hitsSelector() string {
	return `{__name__=~"authorized_hits(_total)?"}`
}

func callsSelector() string {
	return `{__name__=~"authorized_calls(_total)?"}`
}

func limitedSelector() string {
	return `{__name__=~"limited_calls(_total)?"}`
}

func promptSelector() string {
	return `{__name__=~"vllm:prompt_tokens_total"}`
}

func generationSelector() string {
	return `{__name__=~"vllm:generation_tokens_total"}`
}

func increaseBy(selector, by, window string) string {
	return fmt.Sprintf("sum by (%s) (increase(%s[%s]))", by, selector, window)
}

func parseKey(key string) (user, sub, model string) {
	parts := strings.Split(key, "|")
	if len(parts) > 0 {
		user = parts[0]
	}
	if len(parts) > 1 {
		sub = parts[1]
	}
	if len(parts) > 2 {
		model = parts[2]
	}
	return
}

func firstPromErr(errs ...error) string {
	for _, err := range errs {
		if err != nil {
			return err.Error()
		}
	}
	return ""
}
