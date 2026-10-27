package primarytarget

import (
	"context"
	"errors"
	"strings"
	"testing"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type targetRowFixture func(...any) error

func (fixture targetRowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

func TestGetPrimaryTenantTargetReadsOwnerRoutine(t *testing.T) {
	readerPort, err := NewReader(func(
		_ context.Context,
		statement string,
		arguments ...any,
	) Row {
		if len(arguments) != 0 || !strings.Contains(
			statement, "platform.tenancy_get_primary_tenant_target_v1()",
		) {
			t.Fatalf("query = %q %#v", statement, arguments)
		}
		return targetRowFixture(func(destinations ...any) error {
			environmentID := "10000000-0000-4000-8000-000000000001"
			mode := "onpremise"
			tenantID := "50000000-0000-4000-8000-000000000001"
			purpose := "onpremise_primary"
			slug := "harpy"
			*destinations[0].(**string) = &environmentID
			*destinations[1].(**string) = &mode
			*destinations[2].(**string) = &tenantID
			*destinations[3].(**string) = &purpose
			*destinations[4].(**string) = &slug
			*destinations[5].(*bool) = true
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	target, found, err := readerPort.GetPrimaryTenantTarget(context.Background())
	if err != nil || !found || target.CanonicalSlug() != "harpy" ||
		target.Purpose() != tenantapi.PrimaryTenantPurposeOnPremise {
		t.Fatalf("target = %#v, %t, %v", target, found, err)
	}
}

func TestGetPrimaryTenantTargetDistinguishesMissingAndInvalid(t *testing.T) {
	missing, _ := NewReader(func(context.Context, string, ...any) Row {
		return targetRowFixture(func(destinations ...any) error {
			*destinations[5].(*bool) = false
			return nil
		})
	})
	if _, found, err := missing.GetPrimaryTenantTarget(
		context.Background(),
	); err != nil || found {
		t.Fatalf("missing = %t, %v", found, err)
	}
	invalid, _ := NewReader(func(context.Context, string, ...any) Row {
		return targetRowFixture(func(destinations ...any) error {
			environmentID := "10000000-0000-4000-8000-000000000001"
			mode := "saas"
			tenantID := "50000000-0000-4000-8000-000000000001"
			purpose := "onpremise_primary"
			slug := "control"
			*destinations[0].(**string) = &environmentID
			*destinations[1].(**string) = &mode
			*destinations[2].(**string) = &tenantID
			*destinations[3].(**string) = &purpose
			*destinations[4].(**string) = &slug
			*destinations[5].(*bool) = true
			return nil
		})
	})
	if _, _, err := invalid.GetPrimaryTenantTarget(
		context.Background(),
	); !errors.Is(err, errTargetUnavailable) {
		t.Fatalf("invalid error = %v", err)
	}
}

func TestGetPrimaryTenantTargetRejectsUnconfiguredReader(t *testing.T) {
	if _, err := NewReader(nil); !errors.Is(err, errTargetUnavailable) {
		t.Fatalf("constructor error = %v", err)
	}
}
