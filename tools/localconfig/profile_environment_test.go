package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedProfileEnvironmentsSatisfyExactRuntimeCatalog(t *testing.T) {
	set, _ := generatedSetForTest(t)
	tests := []struct {
		file, profile string
	}{
		{"api.env", "api"},
		{"worker.env", "worker"},
		{"migrator.env", "migrator_deploy"},
		{"schema-executor.env", "schema_executor"},
		{"tenant-migration-executor.env", "tenant_migration_executor"},
	}
	catalog := loadRuntimeCatalog(t)
	for _, test := range tests {
		t.Run(test.profile, func(t *testing.T) {
			variables := parseProfileEnvironment(t, set.files[test.file])
			validateProfileCatalog(t, catalog, test.profile, variables)
			if _, inline := variables["HARPY_POSTGRES_DSN"]; inline {
				t.Fatal("PostgreSQL DSN was embedded instead of file-mounted")
			}
		})
	}
}

type runtimeCatalogVariable struct {
	Name        string   `json:"name"`
	Profiles    []string `json:"profiles"`
	Required    bool     `json:"required"`
	FileVariant string   `json:"file_variant"`
}

func loadRuntimeCatalog(t *testing.T) []runtimeCatalogVariable {
	t.Helper()
	var combined []runtimeCatalogVariable
	for _, name := range []string{"config.json", "config.tail.json", "phase14-config.json"} {
		path := filepath.Join("..", "..", "contracts", "runtime", "v1", name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Variables []runtimeCatalogVariable `json:"variables"`
		}
		if err := json.Unmarshal(content, &document); err != nil {
			t.Fatal(err)
		}
		combined = append(combined, document.Variables...)
	}
	return combined
}

func validateProfileCatalog(
	t *testing.T,
	catalog []runtimeCatalogVariable,
	profile string,
	actual map[string]any,
) {
	t.Helper()
	allowed := make(map[string]bool)
	required := make([]runtimeCatalogVariable, 0)
	for _, variable := range catalog {
		if !containsString(variable.Profiles, profile) {
			continue
		}
		allowed[variable.Name] = true
		if variable.FileVariant != "" {
			allowed[variable.FileVariant] = true
		}
		if variable.Required {
			required = append(required, variable)
		}
	}
	for key := range actual {
		if !allowed[key] {
			t.Fatalf("profile contains unknown key %s", key)
		}
	}
	for _, variable := range required {
		_, direct := actual[variable.Name]
		_, fromFile := actual[variable.FileVariant]
		if !direct && (variable.FileVariant == "" || !fromFile) {
			t.Fatalf("profile is missing required key %s", variable.Name)
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func TestGeneratedServiceEnvironmentsUseMountedSecretsOnly(t *testing.T) {
	set, _ := generatedSetForTest(t)
	for _, name := range []string{"api.env", "worker.env"} {
		variables := parseProfileEnvironment(t, set.files[name])
		for key, value := range variables {
			if strings.HasSuffix(key, "_FILE") &&
				!mountedRuntimeFile(value.(string)) {
				t.Fatalf("%s has non-mounted file %s=%s", name, key, value)
			}
		}
	}
	worker := parseProfileEnvironment(t, set.files["worker.env"])
	if _, leaked := worker["DEPLOYMENT_MODE"]; leaked {
		t.Fatal("worker environment contains API-only DEPLOYMENT_MODE")
	}
	if worker["HARPY_IAM_INVITATION_SMTP_USERNAME"] != "harpy-local" {
		t.Fatal("worker and Mailpit SMTP identities differ")
	}
	api := parseProfileEnvironment(t, set.files["api.env"])
	if api["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] !=
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json" ||
		api["HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE"] !=
			"/run/recovery-signing/current/recovery-control-signer.ed25519" {
		t.Fatal("api recovery artifacts do not use the mounted current generation")
	}
}

func mountedRuntimeFile(path string) bool {
	for _, root := range []string{
		secretRoot, "/run/dev-authority/", "/run/recovery-signing/",
	} {
		if strings.HasPrefix(path, root) {
			return true
		}
	}
	return false
}

func TestGeneratedOnPremiseEnvironmentOmitsDeploymentMode(t *testing.T) {
	set, _ := generatedSetForTest(t)
	for _, name := range []string{
		"api.env", "worker.env", "migrator.env", "schema-executor.env",
		"tenant-migration-executor.env",
	} {
		variables := parseProfileEnvironment(t, set.files[name])
		if _, exists := variables["DEPLOYMENT_MODE"]; exists {
			t.Fatalf("%s exposes customer deployment mode configuration", name)
		}
	}
}

func parseProfileEnvironment(t *testing.T, content []byte) map[string]any {
	t.Helper()
	values := make(map[string]any)
	previous := ""
	for _, line := range strings.Split(strings.TrimSuffix(string(content), "\n"), "\n") {
		key, value, found := strings.Cut(line, "=")
		if !found || key == "" || value == "" || key <= previous {
			t.Fatalf("environment is not canonical at %q", line)
		}
		if _, duplicate := values[key]; duplicate {
			t.Fatalf("duplicate environment key %s", key)
		}
		values[key] = value
		previous = key
	}
	return values
}
