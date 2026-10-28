package main

import (
	"encoding/json"
	"testing"

	passwordblocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestGeneratedPasswordBlocklistsMatchManifestCoordinates(t *testing.T) {
	set, _ := generatedSetForTest(t)
	for _, test := range []struct {
		file string
		kind passwordblocklistv1.Kind
	}{
		{"iam-password-common.bin", passwordblocklistv1.KindCommon},
		{"iam-password-compromised.bin", passwordblocklistv1.KindCompromised},
	} {
		artifact, err := passwordblocklistv1.Parse(set.files[test.file], test.kind)
		if err != nil {
			t.Fatalf("parse %s: %v", test.file, err)
		}
		if count, err := artifact.EntryCount(); err != nil || count != 1 {
			t.Fatalf("entry count %s: count=%d err=%v", test.file, count, err)
		}
	}
	var document struct {
		Common, Compromised struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		}
	}
	if err := json.Unmarshal(set.files["iam-password-blocklists.json"], &document); err != nil {
		t.Fatal(err)
	}
	if document.Common.Path != "/run/secrets/iam-password-common.bin" ||
		document.Common.SHA256 != sha256Hex(set.files["iam-password-common.bin"]) ||
		document.Compromised.Path != "/run/secrets/iam-password-compromised.bin" ||
		document.Compromised.SHA256 != sha256Hex(set.files["iam-password-compromised.bin"]) {
		t.Fatal("password blocklist coordinates differ")
	}
}
