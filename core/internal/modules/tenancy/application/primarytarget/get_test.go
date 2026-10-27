package primarytarget

import (
	"context"
	"errors"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type targetReaderFixture func(
	context.Context,
) (tenantapi.PrimaryTenantTarget, bool, error)

func (fixture targetReaderFixture) GetPrimaryTenantTarget(
	ctx context.Context,
) (tenantapi.PrimaryTenantTarget, bool, error) {
	return fixture(ctx)
}

func TestGetPrimaryTenantTargetReturnsPersistedBinding(t *testing.T) {
	target, _ := tenantapi.NewPrimaryTenantTarget(
		"10000000-0000-4000-8000-000000000001",
		tenantapi.DeploymentModeOnPremise,
		"50000000-0000-4000-8000-000000000001",
		tenantapi.PrimaryTenantPurposeOnPremise,
		"harpy",
	)
	query, err := New(targetReaderFixture(func(
		context.Context,
	) (tenantapi.PrimaryTenantTarget, bool, error) {
		return target, true, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	result, err := query.GetPrimaryTenantTarget(context.Background())
	if err != nil || result.TenantID() != target.TenantID() ||
		result.CanonicalSlug() != "harpy" {
		t.Fatalf("target = %#v, %v", result, err)
	}
}

func TestGetPrimaryTenantTargetFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		reader Reader
		want   error
	}{
		{"missing", targetReaderFixture(func(context.Context) (tenantapi.PrimaryTenantTarget, bool, error) {
			return tenantapi.PrimaryTenantTarget{}, false, nil
		}), tenantapi.ErrPrimaryTenantTargetMissing},
		{"invalid", targetReaderFixture(func(context.Context) (tenantapi.PrimaryTenantTarget, bool, error) {
			return tenantapi.PrimaryTenantTarget{}, true, nil
		}), tenantapi.ErrPrimaryTenantTargetUnavailable},
		{"storage", targetReaderFixture(func(context.Context) (tenantapi.PrimaryTenantTarget, bool, error) {
			return tenantapi.PrimaryTenantTarget{}, false, errors.New("storage failed")
		}), tenantapi.ErrPrimaryTenantTargetUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := New(test.reader)
			if err != nil {
				t.Fatal(err)
			}
			_, err = query.GetPrimaryTenantTarget(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestGetPrimaryTenantTargetRejectsMissingDependencies(t *testing.T) {
	if _, err := New(nil); !errors.Is(
		err, tenantapi.ErrPrimaryTenantTargetUnavailable,
	) {
		t.Fatalf("constructor error = %v", err)
	}
}
