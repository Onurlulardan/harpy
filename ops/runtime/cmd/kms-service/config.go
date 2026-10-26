package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	listenKey   = "HARPY_KMS_SERVICE_LISTEN_ADDR"
	certKey     = "HARPY_KMS_SERVICE_TLS_CERT_FILE"
	tlsKeyKey   = "HARPY_KMS_SERVICE_TLS_KEY_FILE"
	masterKey   = "HARPY_KMS_SERVICE_MASTER_KEY_FILE"
	profileKey  = "HARPY_KMS_SERVICE_PROFILE"
	tokenKey    = "HARPY_KMS_SERVICE_ACCESS_TOKEN_SIGNING_KEY_FILE"
	workloadKey = "HARPY_KMS_SERVICE_WORKLOAD_PUBLIC_KEYS_FILE"
)

type runtimeConfig struct {
	listenAddr string
	certFile   string
	tlsKeyFile string
	masterKey  []byte
	profile    string
	tokenKey   []byte
	workloads  map[string]workloadIdentity
}

type environmentLookup func(string) (string, bool)

func loadConfig(lookup environmentLookup) (runtimeConfig, error) {
	listen, err := requiredValue(lookup, listenKey)
	if err != nil {
		return runtimeConfig{}, err
	}
	if _, _, err = net.SplitHostPort(listen); err != nil {
		return runtimeConfig{}, fmt.Errorf("%s is invalid", listenKey)
	}
	cert, err := requiredFile(lookup, certKey)
	if err != nil {
		return runtimeConfig{}, err
	}
	key, err := requiredFile(lookup, tlsKeyKey)
	if err != nil {
		return runtimeConfig{}, err
	}
	if _, err = tls.LoadX509KeyPair(cert, key); err != nil {
		return runtimeConfig{}, errors.New("local KMS TLS key pair is invalid")
	}
	master, err := requiredFile(lookup, masterKey)
	if err != nil {
		return runtimeConfig{}, err
	}
	material, err := os.ReadFile(master)
	if err != nil || len(material) != 32 {
		clear(material)
		return runtimeConfig{}, fmt.Errorf("%s must contain 32 raw bytes", masterKey)
	}
	profile := "local"
	if value, found := lookup(profileKey); found {
		profile = value
	}
	if profile != "local" && profile != "host" {
		clear(material)
		return runtimeConfig{}, fmt.Errorf("%s is invalid", profileKey)
	}
	var subjectKey []byte
	var workloads map[string]workloadIdentity
	if profile == "host" {
		path, keyErr := requiredFile(lookup, tokenKey)
		if keyErr != nil {
			clear(material)
			return runtimeConfig{}, keyErr
		}
		subjectKey, keyErr = os.ReadFile(path)
		if keyErr != nil || len(subjectKey) != 32 {
			clear(material)
			clear(subjectKey)
			return runtimeConfig{}, fmt.Errorf("%s must contain 32 raw bytes", tokenKey)
		}
		registryPath, registryErr := requiredFile(lookup, workloadKey)
		if registryErr != nil {
			clear(material)
			clear(subjectKey)
			return runtimeConfig{}, registryErr
		}
		workloads, registryErr = loadWorkloadRegistry(registryPath)
		if registryErr != nil {
			clear(material)
			clear(subjectKey)
			return runtimeConfig{}, registryErr
		}
	}
	return runtimeConfig{
		listenAddr: listen, certFile: cert, tlsKeyFile: key,
		masterKey: material, profile: profile, tokenKey: subjectKey,
		workloads: workloads,
	}, nil
}

func requiredValue(lookup environmentLookup, key string) (string, error) {
	value, ok := lookup(key)
	if !ok || value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func requiredFile(lookup environmentLookup, key string) (string, error) {
	value, err := requiredValue(lookup, key)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(value) {
		return "", fmt.Errorf("%s must be absolute", key)
	}
	info, err := os.Stat(value)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s must name a regular file", key)
	}
	return value, nil
}
