package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestNotificationsKEKKeyringLoadsCopiesAndRedacts(t *testing.T) {
	loaded, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := loaded.NotificationsKEKKeyring()
	if err != nil {
		t.Fatal(err)
	}
	keys, err := keyring.Keys()
	if err != nil || keys[2] == [32]byte{} {
		t.Fatalf("Keys() = %v, %v", keys, err)
	}
	delete(keys, 2)
	again, _ := keyring.Keys()
	if len(again) != 1 {
		t.Fatal("returned key map aliases keyring state")
	}
	if got := fmt.Sprintf("%v %#v", keyring, keyring); strings.Contains(got, "KSkp") {
		t.Fatalf("keyring leaked: %s", got)
	}
}

func TestNotificationsKEKKeyringRejectsInvalidDocuments(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	valid := `{"version":1,"current":{"revision":1,"key":"` + key +
		`"},"retained":[]}`
	invalid := []string{
		strings.Replace(valid, `"version":1`, `"version":2`, 1),
		strings.Replace(valid, `"retained":[]`, `"retained":[{"revision":1,"key":"`+key+`"}]`, 1),
		strings.Replace(valid, `"key":"`+key+`"`, `"key":"AA"`, 1),
		valid + "\n",
	}
	for _, document := range invalid {
		if _, err := parseNotificationsKEKKeyring([]byte(document)); err == nil {
			t.Fatalf("invalid document accepted: %q", document)
		}
	}
}
