package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// liveCloudPrices is true in production. Tests that must not hit the network
// set it to false so lookupMachineCost uses the embedded catalog.
var liveCloudPrices = true

var priceHTTPClient = &http.Client{Timeout: 8 * time.Second}

func getJSON(url string, dest any) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "maas-finops-bff")
	resp, err := priceHTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return false
	}
	return json.Unmarshal(body, dest) == nil
}
