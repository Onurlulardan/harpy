package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

const manifestName = "internal-saas-profile.manifest.json"

type profileManifest struct {
	Version     int            `json:"version"`
	Profile     string         `json:"profile"`
	InputSHA256 string         `json:"input_sha256"`
	Files       []manifestFile `json:"files"`
}

type manifestFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

func buildManifest(options profileOptions, files map[string][]byte) ([]byte, error) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]manifestFile, 0, len(names))
	for _, name := range names {
		digest := sha256.Sum256(files[name])
		entries = append(entries, manifestFile{
			Name: name, SHA256: hex.EncodeToString(digest[:]),
		})
	}
	input := strings.Join([]string{options.registryURL, options.proposerKMSKeyID,
		options.ackTrustBundle, options.workloadIdentity}, "\x00")
	inputDigest := sha256.Sum256([]byte(input))
	document := profileManifest{Version: 1, Profile: "internal_saas",
		InputSHA256: hex.EncodeToString(inputDigest[:]), Files: entries}
	content, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}
