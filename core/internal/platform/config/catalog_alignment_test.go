package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type catalogDocument struct {
	Variables []catalogVariable `json:"variables"`
}

type catalogVariable struct {
	Name        string   `json:"name"`
	Profiles    []string `json:"profiles"`
	FileVariant string   `json:"file_variant"`
	KeyPurpose  string   `json:"key_purpose"`
}

func TestAllowedKeysMatchPublishedCatalog(t *testing.T) {
	catalog := readCatalog(t)
	profiles := []profile{
		profileAPI, profileWorker, profileMigratorDeploy, profileSchemaExecutor,
		profileClusterBootstrap, profileExtensionAdmin, profileCutoverAdoption,
		profileOperatorInitialize, profileOperatorIdentityReconcile,
		profileOperatorLicenseReconcile,
		profileTenantMigrationExecutor, profileDatabaseBackupExecutor,
		profileDatabaseRestore,
	}
	for _, current := range profiles {
		t.Run(string(current), func(t *testing.T) {
			expected := catalogKeys(catalog, current)
			actual := allowedKeys(current)
			if len(actual) != len(expected) {
				t.Fatalf("key count differs: runtime=%d contract=%d", len(actual), len(expected))
			}
			for key := range expected {
				if !actual[key] {
					t.Fatalf("runtime does not accept contracted key %s", key)
				}
			}
			for key := range actual {
				if !expected[key] {
					t.Fatalf("runtime accepts uncontracted key %s", key)
				}
			}
		})
	}
}

func readCatalog(t *testing.T) catalogDocument {
	t.Helper()
	var combined catalogDocument
	for _, name := range []string{"config.json", "config.tail.json", "phase14-config.json"} {
		path := filepath.Join("..", "..", "..", "..", "contracts", "runtime", "v1", name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var fragment catalogDocument
		if err := json.Unmarshal(content, &fragment); err != nil {
			t.Fatal(err)
		}
		combined.Variables = append(combined.Variables, fragment.Variables...)
	}
	return combined
}

func catalogKeys(catalog catalogDocument, current profile) map[string]bool {
	keys := make(map[string]bool)
	for _, variable := range catalog.Variables {
		for _, contractedProfile := range variable.Profiles {
			if contractedProfile == string(current) {
				keys[variable.Name] = true
				if variable.FileVariant != "" {
					keys[variable.FileVariant] = true
				}
				break
			}
		}
	}
	return keys
}
