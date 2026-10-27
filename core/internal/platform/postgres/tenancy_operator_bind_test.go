package postgres

import (
	"errors"
	"testing"
)

func TestBindTenancyOperatorAllowsOnlyInitializeProfile(t *testing.T) {
	bound, err := BindTenancyOperator(
		tenancyPolicyPool(ProfileOperatorInitialize),
		func(database TenancyOperatorDatabase) (struct{ ready bool }, error) {
			return struct{ ready bool }{ready: database != nil}, nil
		},
	)
	if err != nil || !bound.ready {
		t.Fatalf("operator binding = %#v, %v", bound, err)
	}
	for _, profile := range []Profile{ProfileAPI, ProfileWorker,
		ProfileMigratorDeploy, ProfileTenantMigrationExecutor} {
		called := false
		_, err := BindTenancyOperator(
			tenancyPolicyPool(profile),
			func(TenancyOperatorDatabase) (struct{}, error) {
				called = true
				return struct{}{}, nil
			},
		)
		if !errors.Is(err, ErrTenancyPolicyProfile) || called {
			t.Fatalf("profile %s fence = %v, called=%t", profile, err, called)
		}
	}
}

func TestBindTenancyOperatorRejectsDatabaseEscape(t *testing.T) {
	pool := tenancyPolicyPool(ProfileOperatorInitialize)
	result, err := BindTenancyOperator(
		pool,
		func(database TenancyOperatorDatabase) (any, error) { return database, nil },
	)
	if result != nil || !errors.Is(err, ErrTenancyPolicyBindingInvalid) {
		t.Fatalf("database escape = %#v, %v", result, err)
	}
}
