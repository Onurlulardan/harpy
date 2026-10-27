package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const maximumWorkloadRegistryBytes = 8 * 1024

type workloadIdentity struct {
	subject   string
	publicKey ed25519.PublicKey
}

type workloadRegistryDocument struct {
	Version int `json:"version"`
	Keys    []struct {
		KeyID     string `json:"kid"`
		Subject   string `json:"subject"`
		PublicKey string `json:"public_key"`
	} `json:"keys"`
}

func loadWorkloadRegistry(path string) (map[string]workloadIdentity, error) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= 0 || info.Size() > maximumWorkloadRegistryBytes {
		return nil, fmt.Errorf("%s is invalid", workloadKey)
	}
	document, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workload public keys: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var registry workloadRegistryDocument
	if decoder.Decode(&registry) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		registry.Version != 1 || len(registry.Keys) != 2 {
		return nil, fmt.Errorf("%s is invalid", workloadKey)
	}
	result := make(map[string]workloadIdentity, len(registry.Keys))
	for _, entry := range registry.Keys {
		expectedSubject := "harpy-host-" + entry.KeyID
		publicKey, decodeErr := base64.RawURLEncoding.DecodeString(entry.PublicKey)
		if decodeErr != nil || len(publicKey) != ed25519.PublicKeySize ||
			entry.KeyID != "api" && entry.KeyID != "worker" ||
			entry.Subject != expectedSubject {
			return nil, fmt.Errorf("%s is invalid", workloadKey)
		}
		if _, exists := result[entry.KeyID]; exists {
			return nil, fmt.Errorf("%s contains duplicate keys", workloadKey)
		}
		result[entry.KeyID] = workloadIdentity{
			subject: entry.Subject, publicKey: ed25519.PublicKey(publicKey),
		}
	}
	if _, apiExists := result["api"]; !apiExists {
		return nil, fmt.Errorf("%s is missing api", workloadKey)
	}
	if _, workerExists := result["worker"]; !workerExists {
		return nil, fmt.Errorf("%s is missing worker", workloadKey)
	}
	return result, nil
}
