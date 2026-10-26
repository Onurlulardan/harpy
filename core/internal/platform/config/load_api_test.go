package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadAPIAppliesContractDefaults(t *testing.T) {
	loaded, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PublicListenAddress() != "0.0.0.0:8000" {
		t.Fatalf("unexpected public address: %s", loaded.PublicListenAddress())
	}
	mode, modeErr := loaded.DeploymentMode()
	if modeErr != nil || mode != DeploymentSaaS {
		t.Fatalf("unexpected deployment mode: %s, %v", mode, modeErr)
	}
	if loaded.LongRunning().ManagementListenAddress() != "0.0.0.0:9000" {
		t.Fatalf("unexpected management address: %s", loaded.LongRunning().ManagementListenAddress())
	}
	if loaded.HTTP().ReadHeaderTimeout() != 5*time.Second || loaded.HTTP().UnaryTimeout() != 30*time.Second {
		t.Fatalf("unexpected HTTP defaults: %#v", loaded.HTTP())
	}
	runtimeConfig := loaded.LongRunning().Runtime()
	if runtimeConfig.Environment() != EnvironmentDevelopment || runtimeConfig.LogLevel() != LogInfo {
		t.Fatalf("unexpected runtime defaults")
	}
	if runtimeConfig.Postgres().MaxConnections() != 20 || runtimeConfig.Postgres().MinConnections() != 2 {
		t.Fatalf("unexpected database defaults")
	}
}

func TestLoadAPIDefaultsMissingDeploymentModeToOnPremise(t *testing.T) {
	entries := withoutPhase14APIKeys(validAPIEnvironment())
	entries = append(entries, validAPIOnPremAuthorityEnvironment()...)
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatal(err)
	}
	mode, err := loaded.DeploymentMode()
	if err != nil || mode != DeploymentOnPrem {
		t.Fatalf("deployment mode = %q, %v", mode, err)
	}
}

func TestLoadAPIRejectsPortCollision(t *testing.T) {
	entries := withEnvironment(validAPIEnvironment(),
		"HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:9000",
		"HARPY_MANAGEMENT_LISTEN_ADDR=0.0.0.0:9000",
	)
	_, err := loadAPI(entries)
	if err == nil || !strings.Contains(err.Error(), "different port") {
		t.Fatalf("expected listener collision, got %v", err)
	}
}

func TestLoadAPIRejectsCrossFieldBounds(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		key     string
	}{
		{"min connections", withEnvironment(validAPIEnvironment(), "HARPY_POSTGRES_MAX_CONNS=2", "HARPY_POSTGRES_MIN_CONNS=3"), "MIN_CONNS"},
		{"idle lifetime", withEnvironment(validAPIEnvironment(), "HARPY_POSTGRES_MAX_CONN_LIFETIME=1m", "HARPY_POSTGRES_MAX_CONN_IDLE_TIME=2m"), "IDLE_TIME"},
		{"readiness period", withEnvironment(validAPIEnvironment(), "HARPY_READINESS_TIMEOUT=10s", "HARPY_POSTGRES_HEALTH_CHECK_PERIOD=5s"), "READINESS_TIMEOUT"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadAPI(test.entries)
			if err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("expected %s error, got %v", test.key, err)
			}
		})
	}
}

func TestLoadAPIRejectsNonCanonicalScalars(t *testing.T) {
	tests := []string{
		"HARPY_PUBLIC_LISTEN_ADDR=:8000",
		"HARPY_POSTGRES_MAX_CONNS=+20",
		"HARPY_OTLP_INSECURE=TRUE",
		"HARPY_TRACE_SAMPLE_RATIO=.5",
	}
	for _, invalid := range tests {
		t.Run(invalid, func(t *testing.T) {
			if _, err := loadAPI(withEnvironment(validAPIEnvironment(), invalid)); err == nil {
				t.Fatalf("expected %s to fail", invalid)
			}
		})
	}
}
