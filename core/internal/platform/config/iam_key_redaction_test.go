package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestIAMKeyringsRedactAllFormattingSurfaces(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	loaded, err := loadAPIAt(validAPIEnvironmentAt(now), now)
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, _ := loaded.IAM()
	access, _ := iamConfig.AccessTokenKeyring()
	refresh, _ := iamConfig.RefreshTokenKeyring()
	for name, value := range map[string]any{
		"aggregate":  iamConfig,
		"access":     access,
		"refresh":    refresh,
		"access key": access.active,
		"HMAC key":   refresh.active,
	} {
		t.Run(name, func(t *testing.T) {
			formatted := fmt.Sprintf("%s %v %+v %#v", value, value, value, value)
			if formatted != "[REDACTED] [REDACTED] [REDACTED] [REDACTED]" {
				t.Fatalf("formatting did not redact: %s", formatted)
			}
			document, err := json.Marshal(value)
			if err != nil || string(document) != `"[REDACTED]"` {
				t.Fatalf("JSON did not redact: %s %v", document, err)
			}
		})
	}
	attribute := slog.Any("iam", iamConfig)
	if text := attribute.Value.Resolve().String(); strings.Contains(text, iamFixtureMaterial(10)) {
		t.Fatal("slog value leaked IAM material")
	}
}

func TestZeroIAMConfigsFailClosed(t *testing.T) {
	var api IAMAPI
	var worker IAMWorker
	var access IAMAccessTokenKeyring
	var refresh IAMRefreshKeyring
	var keyID IAMKMSKeyID
	if api.Validate() == nil || worker.Validate() == nil ||
		access.Validate() == nil || refresh.Validate() == nil || keyID.Validate() == nil {
		t.Fatal("zero IAM config was accepted")
	}
}
