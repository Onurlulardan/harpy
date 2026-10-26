package runtimev1

import "testing"

func TestRuntimeIdentityConstants(t *testing.T) {
	t.Parallel()
	if ContractIdentity != "platform-runtime.v1" {
		t.Fatalf("unexpected contract identity %q", ContractIdentity)
	}
	if ContractVersion != 1 {
		t.Fatalf("unexpected contract version %d", ContractVersion)
	}
	if GoVersion != "go1.26.5" {
		t.Fatalf("unexpected Go version %q", GoVersion)
	}
}

func TestServiceValid(t *testing.T) {
	t.Parallel()
	for _, service := range []Service{
		ServiceCoreAPI,
		ServiceCoreMigrator,
		ServiceCoreOperator,
		ServiceCoreWorker,
	} {
		if !service.Valid() {
			t.Errorf("expected service %q to be valid", service)
		}
	}
	if Service("api").Valid() {
		t.Fatal("unexpected service accepted")
	}
}

func TestEnvironmentValid(t *testing.T) {
	t.Parallel()
	for _, environment := range []Environment{
		EnvironmentDevelopment,
		EnvironmentProduction,
		EnvironmentStaging,
		EnvironmentTest,
	} {
		if !environment.Valid() {
			t.Errorf("expected environment %q to be valid", environment)
		}
	}
	if Environment("local").Valid() {
		t.Fatal("unexpected environment accepted")
	}
}
