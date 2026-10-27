package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestTenantStoreFacadeRejectsNil(t *testing.T) {
	if (*Store)(nil).check(context.Background()) == nil {
		t.Fatal("nil store accepted")
	}
}

func TestTenantStoreFacadeDefersSchemaIdentityToPlatformBinding(t *testing.T) {
	for _, schema := range []string{
		"t_abcdefghijklmnopqrstuvwxyz",
		"tenant_harpy",
	} {
		database := databaseFacade{currentSchema: func(context.Context) (string, error) {
			return schema, nil
		}}
		if !database.platformBindingAvailable(context.Background()) {
			t.Fatalf("platform-validated schema %q rejected", schema)
		}
	}
	database := databaseFacade{currentSchema: func(context.Context) (string, error) {
		return "", errors.New("binding rejected")
	}}
	if database.platformBindingAvailable(context.Background()) {
		t.Fatal("platform binding error accepted")
	}
}
