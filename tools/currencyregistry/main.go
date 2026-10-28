// Command currencyregistry records fresh currency provenance.
// It does not write a v1_adopted SQL file.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

const (
	freshGo  = "core/internal/modules/billing/domain/money/currency_registry_gen.go"
	freshSQL = "core/internal/modules/billing/migrations/platform/fresh/" +
		"20260717010012_seed_billing_currency_exponents_v1.fresh.sql"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		exit(err)
	}
	goPath := filepath.Join(root, filepath.FromSlash(freshGo))
	inputDigest, err := normalizedInputDigest(goPath)
	if err != nil {
		exit(err)
	}
	goHash, err := fileSHA256(goPath)
	if err != nil {
		exit(err)
	}
	sqlHash, err := fileSHA256(filepath.Join(root, filepath.FromSlash(freshSQL)))
	if err != nil {
		exit(err)
	}
	if err = writeManifest(root, []generatedFile{
		{Path: freshGo, Generator: "go run ./tools/currencyregistry",
			InputSHA256: inputDigest, OutputSHA256: goHash},
		{Path: freshSQL, Generator: "go run ./tools/currencyregistry",
			InputSHA256: inputDigest, OutputSHA256: sqlHash},
	}); err != nil {
		exit(err)
	}
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

type generatedFile struct {
	Path         string `json:"path"`
	Generator    string `json:"generator"`
	InputSHA256  string `json:"input_sha256"`
	OutputSHA256 string `json:"output_sha256"`
}

func normalizedInputDigest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	const marker = "currencyNormalizedInputSHA256"
	text := string(data)
	index := strings.Index(text, marker)
	if index < 0 {
		return "", os.ErrNotExist
	}
	quoted := text[index+len(marker):]
	start := strings.Index(quoted, `"`)
	end := strings.Index(quoted[start+1:], `"`)
	if start < 0 || end != 64 {
		return "", os.ErrInvalid
	}
	return quoted[start+1 : start+1+end], nil
}

func exit(err error) {
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}
