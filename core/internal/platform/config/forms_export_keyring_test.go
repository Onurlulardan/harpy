package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestFCARTKEKKeyringParsesCanonicalDocument(t *testing.T) {
	keyring, err := parseFCARTKEKKeyring([]byte(validFormsExportKeyringJSON()))
	if err != nil || keyring.Validate() != nil {
		t.Fatalf("valid keyring rejected: %v", err)
	}
	current, _ := keyring.Current()
	id, _ := current.ID()
	material, _ := current.AES256Key()
	retained, _ := keyring.Retained()
	if id != "forms-export-2026-07" || material[0] != 71 || len(retained) != 0 {
		t.Fatal("keyring values changed")
	}
}

func TestFCARTKEKKeyringRejectsNoncanonicalOrUnsafeDocuments(t *testing.T) {
	valid := validFormsExportKeyringJSON()
	zero := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	material := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{72}, 32))
	retained := fmt.Sprintf(`{"kid":"old-%d","key":"%s"}`, 1, material)
	eight := strings.TrimSuffix(valid, "[]}") + "[" + strings.Repeat(retained+",", 7) + retained + "]}"
	tests := []string{
		strings.Replace(valid, `{"version":1`, `{"version":1,"version":1`, 1),
		strings.Replace(valid, `,"current"`, `, "current"`, 1),
		strings.Replace(valid, `,"retained"`, `,"unknown":true,"retained"`, 1),
		strings.Replace(valid, base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{71}, 32)), zero, 1),
		strings.Replace(valid, `"retained":[]`, `"retained":[{"kid":"forms-export-2026-07","key":"`+material+`"}]`, 1),
		eight,
	}
	for index, document := range tests {
		if _, err := parseFCARTKEKKeyring([]byte(document)); err == nil {
			t.Fatalf("invalid keyring %d accepted", index)
		}
	}
}

func TestFormsExportSecretsAreRedacted(t *testing.T) {
	keyring, _ := parseFCARTKEKKeyring([]byte(validFormsExportKeyringJSON()))
	current, _ := keyring.Current()
	credentials := FormsExportS3Credentials{source: FormsExportStaticFiles,
		accessKeyID: Secret{value: "access-secret"}, secretKey: Secret{value: "secret-secret"}, valid: true}
	storage := FormsExportS3{endpoint: "https://s3.invalid", credentials: credentials}
	artifacts := FormsExportArtifacts{storage: storage, keyring: keyring}
	output := fmt.Sprintf("%v %#v %s %v %#v %s %v %#v %s",
		keyring, keyring, keyring, current, current, current,
		credentials, credentials, credentials)
	output += fmt.Sprintf(" %+v %#v %+v %#v", storage, storage, artifacts, artifacts)
	if strings.Contains(output, "access-secret") || strings.Contains(output, "secret-secret") ||
		strings.Contains(output, base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{71}, 32))) {
		t.Fatalf("secret leaked through formatting: %s", output)
	}
	if value := slog.Any("keyring", keyring).Value.Resolve().String(); strings.Contains(value, "forms-export") {
		t.Fatalf("secret leaked through slog: %s", value)
	}
}
