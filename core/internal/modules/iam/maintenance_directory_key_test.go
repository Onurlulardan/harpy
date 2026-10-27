package iam

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestMaintenanceDirectoryKeyRedactsEverySerializationSurface(t *testing.T) {
	material := [32]byte{}
	for index := range material {
		material[index] = byte(index + 1)
	}
	keyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{22}, 12))
	key, err := NewMaintenanceDirectoryKey(keyID, material, true)
	if err != nil {
		t.Fatalf("new key: %v", err)
	}
	encoded, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	var log bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&log, nil))
	logger.Info("key", "value", key)
	outputs := []string{
		fmt.Sprint(key), fmt.Sprintf("%+v", key), fmt.Sprintf("%#v", key),
		fmt.Sprintf("%s", key), fmt.Sprintf("%q", key), fmt.Sprintf("%x", key),
		string(encoded), log.String(),
	}
	secret := hex.EncodeToString(material[:])
	for _, output := range outputs {
		if strings.Contains(output, secret) || !strings.Contains(output, "REDACTED") {
			t.Fatalf("unsafe key serialization: %q", output)
		}
	}
}

func TestMaintenanceDirectoryKeyRejectsInvalidMaterial(t *testing.T) {
	keyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{22}, 12))
	if _, err := NewMaintenanceDirectoryKey(keyID, [32]byte{}, true); err == nil {
		t.Fatal("zero material was accepted")
	}
	material := [32]byte{1}
	if _, err := NewMaintenanceDirectoryKey("INVALID", material, true); err == nil {
		t.Fatal("invalid key ID was accepted")
	}
}
