package deploymentpolicy

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type policyRowFixture func(...any) error

func (fixture policyRowFixture) Scan(destinations ...any) error {
	return fixture(destinations...)
}

func TestGetDeploymentPolicyReadsOwnerRoutine(t *testing.T) {
	environmentID := "10000000-0000-4000-8000-000000000001"
	createdAt := time.Date(2026, 7, 22, 9, 30, 0, 0, time.UTC)
	reader, err := NewPolicyReader(func(
		_ context.Context,
		statement string,
		arguments ...any,
	) Row {
		if len(arguments) != 0 || !strings.Contains(
			statement, "platform.tenancy_get_deployment_policy_v1()",
		) {
			t.Fatalf("query = %q %#v", statement, arguments)
		}
		return policyRowFixture(func(destinations ...any) error {
			*destinations[0].(**string) = &environmentID
			mode := "onpremise"
			*destinations[1].(**string) = &mode
			zero := int64(0)
			*destinations[2].(**int64) = &zero
			*destinations[3].(**time.Time) = &createdAt
			*destinations[4].(*bool) = true
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	policy, found, err := reader.GetDeploymentPolicy(context.Background())
	if err != nil || !found || policy.EnvironmentID() != environmentID ||
		policy.Mode() != tenancyapi.DeploymentModeOnPremise {
		t.Fatalf("policy = %#v, %t, %v", policy, found, err)
	}
}

func TestGetDeploymentPolicyDistinguishesMissingAndInvalid(t *testing.T) {
	missing, _ := NewPolicyReader(func(
		context.Context, string, ...any,
	) Row {
		return policyRowFixture(func(destinations ...any) error {
			*destinations[4].(*bool) = false
			return nil
		})
	})
	if _, found, err := missing.GetDeploymentPolicy(context.Background()); err != nil || found {
		t.Fatalf("missing = %t, %v", found, err)
	}
	invalid, _ := NewPolicyReader(func(
		context.Context, string, ...any,
	) Row {
		return policyRowFixture(func(destinations ...any) error {
			environment := "00000000-0000-0000-0000-000000000000"
			mode := "saas"
			createdAt := time.Now().UTC().Truncate(time.Microsecond)
			*destinations[0].(**string) = &environment
			*destinations[1].(**string) = &mode
			*destinations[3].(**time.Time) = &createdAt
			*destinations[4].(*bool) = true
			return nil
		})
	})
	if _, _, err := invalid.GetDeploymentPolicy(context.Background()); !errors.Is(
		err, errPolicyUnavailable,
	) {
		t.Fatalf("invalid error = %v", err)
	}
}

func TestGetDeploymentPolicyRejectsUnconfiguredReader(t *testing.T) {
	if _, err := NewPolicyReader(nil); !errors.Is(err, errPolicyUnavailable) {
		t.Fatalf("constructor error = %v", err)
	}
	var reader *policyReader
	if _, _, err := reader.GetDeploymentPolicy(context.Background()); !errors.Is(
		err, errPolicyUnavailable,
	) {
		t.Fatalf("nil reader error = %v", err)
	}
}
