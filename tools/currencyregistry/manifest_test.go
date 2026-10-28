package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestRewriteTwiceIsByteIdentical(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "contracts"), 0o755); err != nil {
		t.Fatal(err)
	}
	const (
		keptSpace = `"generator":  "keep-bytes"`
		adopted   = "core/internal/modules/billing/migrations/platform/v1_adopted/old.v1_adopted.sql"
		staleGo   = "1111111111111111111111111111111111111111111111111111111111111111"
		staleSQL  = "2222222222222222222222222222222222222222222222222222222222222222"
		nextGo    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		nextSQL   = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		input     = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	)
	body := strings.Join([]string{
		"{",
		"  \"version\": 1,",
		"  \"files\": [",
		"    {",
		"      \"path\": \"contracts/http/preserve-me.openapi.yaml\",",
		"      " + keptSpace + ",",
		"      \"input_sha256\": \"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd\",",
		"      \"output_sha256\": \"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee\"",
		"    },",
		"    {",
		"      \"path\": \"" + adopted + "\",",
		"      \"generator\": \"historical\",",
		"      \"input_sha256\": \"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\",",
		"      \"output_sha256\": \"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\"",
		"    },",
		"    {\"path\":\"" + freshGo + "\",\"generator\":\"go run ./tools/currencyregistry\",\"input_sha256\":\"" + input + "\",\"output_sha256\":\"" + staleGo + "\"},",
		"    {\"path\":\"" + freshSQL + "\",\"generator\":\"go run ./tools/currencyregistry\",\"input_sha256\":\"" + input + "\",\"output_sha256\":\"" + staleSQL + "\"}",
		"  ]",
		"}",
		"",
	}, "\r\n")
	path := filepath.Join(root, "contracts", "generated.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	replacements := []generatedFile{
		{Path: freshGo, Generator: "go run ./tools/currencyregistry", InputSHA256: input, OutputSHA256: nextGo},
		{Path: freshSQL, Generator: "go run ./tools/currencyregistry", InputSHA256: input, OutputSHA256: nextSQL},
	}
	if err := writeManifest(root, replacements); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeManifest(root, replacements); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("second manifest rewrite changed bytes")
	}
	if !bytes.Contains(second, []byte(keptSpace+",\r\n")) {
		t.Fatal("unrelated record bytes were rewritten")
	}
	if !bytes.Contains(second, []byte("\"generator\": \"historical\",\r\n")) {
		t.Fatal("non-currency record bytes were rewritten")
	}
	if !bytes.Contains(second, []byte(nextGo)) || !bytes.Contains(second, []byte(nextSQL)) {
		t.Fatal("currency provenance was not recorded")
	}
	if bytes.Contains(second, []byte(staleGo)) || bytes.Contains(second, []byte(staleSQL)) {
		t.Fatal("stale currency provenance remained")
	}
}
