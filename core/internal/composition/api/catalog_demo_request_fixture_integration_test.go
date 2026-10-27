//go:build integration

package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const catalogDemoPackageID = "e3000000-0000-0000-0000-000000000009"
const catalogDemoReleaseID = "019c3300-0000-7000-8000-000000000001"
const catalogDemoComponentID = "018f0000-0004-7000-8000-000000000009"
const catalogDemoFormsSHA = "9cbf20cfe4637cceb549d772990f5b87454bfa2edec4b73b0a97a4081190d220"

type catalogDemoOwnerPayloadGroup struct {
	Payloads []struct {
		ComponentID   string          `json:"component_id"`
		ContentSHA256 string          `json:"content_sha256"`
		Contract      string          `json:"contract"`
		Document      json.RawMessage `json:"document"`
		Owner         string          `json:"owner"`
	} `json:"payloads"`
}

func catalogDemoFormsRequest(t *testing.T, root string, owner any) any {
	t.Helper()
	path := filepath.Join(root, "tools", "internal", "firstpartycatalog",
		"testdata", "project", "artifacts", "owner-forms.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var group catalogDemoOwnerPayloadGroup
	if err = json.Unmarshal(content, &group); err != nil {
		t.Fatalf("decode Project Forms owner artifact: %v", err)
	}
	for _, payload := range group.Payloads {
		var identity struct {
			FormID string `json:"form_id"`
			Slug   string `json:"slug"`
		}
		if json.Unmarshal(payload.Document, &identity) != nil ||
			identity.Slug != "project-type" {
			continue
		}
		sum := sha256.Sum256(payload.Document)
		digest := hex.EncodeToString(sum[:])
		if identity.FormID != catalogDemoProjectTypeID ||
			payload.ComponentID != catalogDemoProjectTypeID ||
			payload.Owner != "forms" ||
			payload.Contract != "forms.catalog-definition.v1" ||
			payload.ContentSHA256 != catalogDemoFormsSHA ||
			digest != payload.ContentSHA256 {
			t.Fatalf("Project Forms owner artifact identity differs: %+v digest=%s",
				payload, digest)
		}
		return catalogDemoBuildRequest(t, owner, catalogDemoRequestSpec{
			componentID: payload.ComponentID, owner: payload.Owner,
			contract: payload.Contract, contentSHA256: payload.ContentSHA256,
			document: payload.Document, idempotencyKey: "catalog-forms-project-pg18",
		})
	}
	t.Fatal("Project Forms owner artifact does not contain project-type")
	return nil
}

func catalogDemoCatalogRequest(t *testing.T, owner any) any {
	t.Helper()
	document := []byte(fmt.Sprintf(
		`{"component_id":"%s","contract_version":"catalog.demo-data.v1",`+
			`"demo_forms":[{"legacy_form_config_id":`+
			`"e4000000-0000-0000-0000-000000000561","rows":[`+
			`{"description":"%s","id":"%s","type_name":"Forward Project"},`+
			`{"description":"%s","id":"%s","type_name":"Target Project"}],`+
			`"slug":"project-type"}],"excluded_form_slugs":[],"module_slug":`+
			`"project","source_contract":"harpy-v1.enterprise-seed.project"}`,
		catalogDemoComponentID, catalogDemoLegacyB, catalogDemoLegacyA,
		catalogDemoExternal, catalogDemoLegacyB,
	))
	sum := sha256.Sum256(document)
	return catalogDemoBuildRequest(t, owner, catalogDemoRequestSpec{
		componentID: catalogDemoComponentID, owner: "catalog",
		contract:      "catalog.demo-data.v1",
		contentSHA256: hex.EncodeToString(sum[:]), document: document,
		idempotencyKey: "catalog-demo-project-pg18-forward-reference",
	})
}

type catalogDemoRequestSpec struct {
	packageID      string
	packageSlug    string
	packageVersion string
	releaseID      string
	componentID    string
	owner          string
	contract       string
	contentSHA256  string
	document       []byte
	idempotencyKey string
}

func catalogDemoBuildRequest(
	t *testing.T, owner any, spec catalogDemoRequestSpec,
) any {
	t.Helper()
	method := reflect.ValueOf(owner).MethodByName("Provision")
	if !method.IsValid() || method.Type().NumIn() != 3 {
		t.Fatalf("Catalog component owner %s does not expose Provision", spec.owner)
	}
	if spec.packageID == "" {
		spec.packageID = catalogDemoPackageID
	}
	if spec.packageSlug == "" {
		spec.packageSlug = "project"
	}
	if spec.releaseID == "" {
		spec.releaseID = catalogDemoReleaseID
	}
	if spec.packageVersion == "" {
		spec.packageVersion = "1.0.0"
	}
	request := reflect.New(method.Type().In(2)).Elem()
	setCatalogDemoString(t, request, "PackageID", spec.packageID)
	setCatalogDemoString(t, request, "PackageSlug", spec.packageSlug)
	setCatalogDemoString(t, request, "PackageVersion", spec.packageVersion)
	setCatalogDemoString(t, request, "ReleaseID", spec.releaseID)
	setCatalogDemoString(t, request, "IdempotencyKey", spec.idempotencyKey)
	request.FieldByName("Document").SetBytes(append([]byte(nil), spec.document...))
	component := request.FieldByName("Component")
	setCatalogDemoString(t, component, "ID", spec.componentID)
	setCatalogDemoString(t, component, "Owner", spec.owner)
	component.FieldByName("Generation").SetInt(7)
	setCatalogDemoString(t, component, "Contract", spec.contract)
	setCatalogDemoString(t, component, "ContentSHA256", spec.contentSHA256)
	return request.Interface()
}

func setCatalogDemoString(
	t *testing.T, value reflect.Value, field string, content string,
) {
	t.Helper()
	target := value.FieldByName(field)
	if !target.IsValid() || !target.CanSet() || target.Kind() != reflect.String {
		t.Fatalf("Catalog demo request field %s is unavailable", field)
	}
	target.SetString(content)
}

func catalogDemoUUIDv7(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' || value[14] != '7' ||
		!strings.ContainsRune("89ab", rune(value[19])) {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func catalogDemoRepositoryRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err = os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatalf("backend-v2 repository root not found from %s", current)
		}
		current = parent
	}
}
