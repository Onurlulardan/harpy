package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	listen := value("S3_OWNER_PROXY_LISTEN_ADDR", "0.0.0.0:9000")
	owner := value("S3_OWNER_EXPECTED_BUCKET_OWNER", "")
	certificate := value("S3_OWNER_PROXY_TLS_CERT_FILE", "")
	privateKey := value("S3_OWNER_PROXY_TLS_KEY_FILE", "")
	upstream, err := url.Parse(value("S3_OWNER_PROXY_UPSTREAM_URL", "http://minio:9000"))
	if err != nil || upstream.Scheme != "http" || upstream.Host == "" || owner == "" ||
		(certificate == "") != (privateKey == "") {
		log.Fatal("invalid S3 owner proxy configuration")
	}
	handler, err := newOwnerProxy(upstream, owner)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:              listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("S3 owner proxy listening on %s", listen)
	if certificate != "" {
		err = server.ListenAndServeTLS(certificate, privateKey)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("serve S3 owner proxy: %w", err))
	}
}

func value(name, fallback string) string {
	if result := os.Getenv(name); result != "" {
		return result
	}
	return fallback
}
