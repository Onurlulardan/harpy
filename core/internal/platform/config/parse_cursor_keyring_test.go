package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestParseCursorKeyringAcceptsExactKeys(t *testing.T) {
	keyring, err := parseCursorKeyring(Secret{value: validCursorKeyringJSON()})
	if err != nil {
		t.Fatal(err)
	}
	current, err := keyring.Current()
	if err != nil {
		t.Fatal(err)
	}
	previous, err := keyring.Previous()
	if err != nil {
		t.Fatal(err)
	}
	currentKID, err := current.KID()
	if err != nil {
		t.Fatal(err)
	}
	previousKID, err := previous.KID()
	if err != nil {
		t.Fatal(err)
	}
	if currentKID != "current-1" || previousKID != "previous-1" {
		t.Fatalf("unexpected key IDs: %s %s", currentKID, previousKID)
	}
	material, err := current.AES256Key()
	if err != nil {
		t.Fatal(err)
	}
	var expected [cursorKeySize]byte
	for index := range expected {
		expected[index] = 1
	}
	if material != expected {
		t.Fatal("current key material differs")
	}
}

func TestParseCursorKeyringRejectsContractDrift(t *testing.T) {
	valid := validCursorKeyringJSON()
	tests := []struct {
		name  string
		value string
	}{
		{"version", strings.Replace(valid, `"version":1`, `"version":2`, 1)},
		{"missing previous", fmt.Sprintf(`{"version":1,"current":{"kid":"current-1","key":"%s"}}`, cursorTestKey(1))},
		{"same kid", strings.Replace(valid, `"previous-1"`, `"current-1"`, 1)},
		{"uppercase kid", strings.Replace(valid, `"current-1"`, `"Current-1"`, 1)},
		{"long kid", strings.Replace(valid, `"current-1"`, `"abcdefghijklmnopqrstuvwxyz1234567"`, 1)},
		{"padded key", cursorTestJSON("current-1", cursorTestKey(1)+"=", "previous-1", cursorTestKey(2))},
		{"short key", cursorTestJSON("current-1", cursorTestKeyBytes(31), "previous-1", cursorTestKey(2))},
		{"invalid key", cursorTestJSON("current-1", "must-not-leak", "previous-1", cursorTestKey(2))},
		{"unknown root", strings.TrimSuffix(valid, "}") + `,"extra":true}`},
		{"unknown key member", strings.Replace(valid, `"kid":"current-1"`, `"kid":"current-1","extra":true`, 1)},
		{"wrong member case", strings.Replace(valid, `"current":`, `"Current":`, 1)},
		{"duplicate root", strings.Replace(valid, `"version":1`, `"version":1,"version":1`, 1)},
		{"duplicate nested", strings.Replace(valid, `"kid":"current-1"`, `"kid":"current-1","kid":"other"`, 1)},
		{"trailing JSON", valid + `{}`},
		{"invalid UTF-8", valid[:len(valid)-1] + string([]byte{0xff}) + "}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseCursorKeyring(Secret{value: test.value})
			if err == nil {
				t.Fatal("invalid cursor keyring accepted")
			}
			if strings.Contains(err.Error(), "must-not-leak") || strings.Contains(err.Error(), cursorTestKey(1)) {
				t.Fatalf("cursor keyring error leaked secret: %v", err)
			}
		})
	}
}

func cursorTestJSON(currentKID, currentKey, previousKID, previousKey string) string {
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"%s","key":"%s"},"previous":{"kid":"%s","key":"%s"}}`,
		currentKID, currentKey, previousKID, previousKey,
	)
}

func cursorTestKey(value byte) string {
	return cursorTestKeyBytesWithValue(cursorKeySize, value)
}

func cursorTestKeyBytes(size int) string {
	return cursorTestKeyBytesWithValue(size, 1)
}

func cursorTestKeyBytesWithValue(size int, value byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, size))
}
