package main

import (
	"bytes"
	"testing"
)

func TestGeneratedNotificationDocumentsUseRuntimeCanonicalOrder(t *testing.T) {
	set, _ := generatedSetForTest(t)
	tests := []struct {
		file   string
		prefix string
	}{
		{"notifications-consumer-manifest.json", `{"version":1,"consumer":"frontend-v2"`},
		{"notifications-consumer-keyring.json", `{"version":1,"keys":[{"key_id":`},
		{"notifications-kek-keyring.json", `{"version":1,"current":{"revision":1,"key":`},
	}
	for _, test := range tests {
		if !bytes.HasPrefix(set.files[test.file], []byte(test.prefix)) {
			t.Fatalf("%s does not use runtime canonical field order", test.file)
		}
	}
}
