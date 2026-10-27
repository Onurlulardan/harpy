package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func importSourceKeyringJSON(kid string, marker byte) string {
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":%q,"key":%q},"retained":[]}`, kid, key)
}

func TestFormsImportSourceKeyringIsCanonicalAndRedacted(t *testing.T) {
	raw := importSourceKeyringJSON("forms-import-source-current", 91)
	keyring, err := parseFormsImportSourceKEKKeyring([]byte(raw))
	if err != nil || keyring.Validate() != nil {
		t.Fatalf("parse import-source keyring: %v", err)
	}
	if strings.Contains(fmt.Sprintf("%v %#v", keyring, keyring), raw) {
		t.Fatal("import-source keyring formatting leaked material")
	}
	for _, invalid := range []string{raw + "\n",
		strings.Replace(raw, `"retained":[]`, `"retained":null`, 1),
		strings.Replace(raw, `"version":1`, `"version":2`, 1)} {
		if _, err := parseFormsImportSourceKEKKeyring([]byte(invalid)); err == nil {
			t.Fatalf("invalid import-source keyring accepted: %q", invalid)
		}
	}
}

func TestFormsImportSourceRejectsExportKIDOrMaterialReuse(t *testing.T) {
	tests := []string{importSourceKeyringJSON("forms-export-2026-07", 91),
		importSourceKeyringJSON("different-import-kid", 71)}
	for _, document := range tests {
		directory := t.TempDir()
		path := filepath.Join(directory, "keyring.json")
		if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
			t.Fatal(err)
		}
		entries := withEnvironment(validAPIEnvironment(),
			formsImportSourceFCARTKeyringFileKey+"="+path)
		if _, err := loadAPI(entries); err == nil ||
			!strings.Contains(err.Error(), "must not share") {
			t.Fatalf("cross-domain key reuse error = %v", err)
		}
	}
}
