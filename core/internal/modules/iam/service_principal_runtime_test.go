package iam

import (
	"testing"
	"time"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func TestServicePrincipalRuntimeRequiresSealedCatalogsAndCapabilityManifest(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	key, err := serviceapp.ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	definitions := []ServicePrincipalCapabilityDefinition{{
		Key: key, Owner: "forms", ResourceKind: "record",
		TenantScope: true, ResourceScope: true,
	}}
	runtime, err := NewServicePrincipalRuntime(catalogs, definitions, time.Now)
	if err != nil || !runtime.Valid() {
		t.Fatalf("runtime = %#v, %v", runtime, err)
	}
	if _, err := NewServicePrincipalRuntime(
		platformasync.WriteCatalogs{}, definitions, time.Now,
	); err == nil {
		t.Fatal("unsealed write catalogs were accepted")
	}
	if _, err := NewServicePrincipalRuntime(catalogs, definitions, nil); err == nil {
		t.Fatal("missing clock was accepted")
	}
	if _, err := NewServicePrincipalRuntime(catalogs, nil, time.Now); err == nil {
		t.Fatal("empty capability manifest was accepted")
	}
}
