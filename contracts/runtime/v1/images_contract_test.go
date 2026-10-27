package runtimev1

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

type imageContract struct {
	Version        int              `json:"version"`
	SourcePostgres postgresContract `json:"source_postgres"`
	Postgres       postgresContract `json:"postgres"`
}

type postgresContract struct {
	EngineVersion string            `json:"engine_version"`
	Repository    string            `json:"repository"`
	Tag           string            `json:"tag"`
	Digest        string            `json:"oci_index_digest"`
	Encoding      string            `json:"encoding"`
	Locale        string            `json:"locale"`
	DataChecksums bool              `json:"data_checksums"`
	Extensions    map[string]string `json:"extensions"`
	Platforms     []string          `json:"platforms"`
}

func TestPostgresImageContractIsPinnedAndUsedByIntegration(t *testing.T) {
	content, err := os.ReadFile("images.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract imageContract
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		t.Fatalf("decode image contract: %v", err)
	}
	if contract.Version != 3 || contract.Postgres.EngineVersion != PostgresEngineVersion {
		t.Fatalf("unexpected image contract version: %#v", contract)
	}
	if contract.SourcePostgres.EngineVersion != SourcePostgresEngineVersion ||
		contract.SourcePostgres.Digest != SourcePostgresImageDigest ||
		contract.SourcePostgres.Extensions["vector"] != VectorExtensionVersion {
		t.Fatalf("unexpected source PostgreSQL contract: %#v", contract.SourcePostgres)
	}
	if contract.Postgres.Digest != PostgresImageDigest ||
		contract.Postgres.Extensions["vector"] != VectorExtensionVersion {
		t.Fatalf("unexpected vector version: %#v", contract.Postgres.Extensions)
	}
	if contract.SourcePostgres.Encoding != PostgresEncoding ||
		contract.Postgres.Encoding != PostgresEncoding ||
		contract.SourcePostgres.Locale != PostgresCollation ||
		contract.Postgres.Locale != PostgresCollation ||
		contract.SourcePostgres.DataChecksums ||
		!contract.Postgres.DataChecksums {
		t.Fatalf("source/target locale contract mismatch: %#v", contract)
	}
	if !regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(contract.Postgres.Digest) {
		t.Fatalf("invalid OCI digest %q", contract.Postgres.Digest)
	}
	compose, err := os.ReadFile("../../../ops/integration/compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	image := contract.Postgres.Repository + ":" + contract.Postgres.Tag + "@" + contract.Postgres.Digest
	if !strings.Contains(string(compose), "image: "+image) {
		t.Fatalf("integration runtime does not use pinned image %q", image)
	}
	if !strings.Contains(string(compose), "--locale="+contract.Postgres.Locale) {
		t.Fatalf("integration runtime does not use locale %q", contract.Postgres.Locale)
	}
}
