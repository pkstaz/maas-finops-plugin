package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rh-ai-community-plugins/maas-finops-plugin/pkg/server"
)

func main() {
	port := getenv("PORT", "3000")
	dist := os.Getenv("PLUGIN_DIST_DIR")
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	srv, err := server.New(dist)
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	if certFile != "" && keyFile != "" {
		absCert, _ := filepath.Abs(certFile)
		absKey, _ := filepath.Abs(keyFile)
		httpServer.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		log.Printf("listening on :%s (tls) dist=%s", port, dist)
		log.Fatal(httpServer.ListenAndServeTLS(absCert, absKey))
	}

	log.Printf("listening on :%s (http) dist=%s", port, dist)
	log.Fatal(httpServer.ListenAndServe())
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
