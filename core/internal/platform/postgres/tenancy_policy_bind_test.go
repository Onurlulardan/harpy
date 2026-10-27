package postgres

import (
	"errors"
	"testing"
)

func TestBindTenancyPolicyReaderAllowsRuntimeProfiles(t *testing.T) {
	t.Parallel()
	for _, profile := range []Profile{ProfileAPI, ProfileWorker} {
		pool := tenancyPolicyPool(profile)
		reader, err := BindTenancyPolicyReader(
			pool,
			func(query TenancyPolicyQueryRow) (struct{ bound bool }, error) {
				return struct{ bound bool }{bound: query != nil}, nil
			},
		)
		if err != nil || !reader.bound {
			t.Fatalf("profile %s reader = %#v, %v", profile, reader, err)
		}
	}
}

func TestBindTenancyPolicyReaderRejectsOtherProfiles(t *testing.T) {
	t.Parallel()
	for _, profile := range []Profile{ProfileMigratorDeploy, ProfileSchemaExecutor} {
		called := false
		_, err := BindTenancyPolicyReader(
			tenancyPolicyPool(profile),
			func(TenancyPolicyQueryRow) (struct{}, error) {
				called = true
				return struct{}{}, nil
			},
		)
		if !errors.Is(err, ErrTenancyPolicyProfile) || called {
			t.Fatalf("profile %s fence = %v, called = %t", profile, err, called)
		}
	}
}

func TestBindTenancyPolicyReaderRejectsEscapesAndInvalidInput(t *testing.T) {
	t.Parallel()
	pool := tenancyPolicyPool(ProfileAPI)
	if _, err := BindTenancyPolicyReader[any](pool, nil); !errors.Is(
		err, ErrTenancyPolicyBindingInvalid,
	) {
		t.Fatalf("nil factory error = %v", err)
	}
	result, err := BindTenancyPolicyReader(
		pool,
		func(query TenancyPolicyQueryRow) (any, error) { return query, nil },
	)
	if result != nil || !errors.Is(err, ErrTenancyPolicyBindingInvalid) {
		t.Fatalf("executor escape = %#v, %v", result, err)
	}
	want := errors.New("factory failed")
	_, err = BindTenancyPolicyReader(
		pool,
		func(TenancyPolicyQueryRow) (struct{}, error) { return struct{}{}, want },
	)
	if !errors.Is(err, want) {
		t.Fatalf("factory error = %v", err)
	}
}

func tenancyPolicyPool(profile Profile) *Pool {
	return &Pool{profile: profile, contract: profileContract{runtime: true}}
}
