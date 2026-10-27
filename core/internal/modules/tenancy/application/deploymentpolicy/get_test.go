package deploymentpolicy

import (
	"context"
	"errors"
	"testing"
	"time"

	tenancyapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

type policyReaderFixture func(
	context.Context,
) (tenancyapi.DeploymentPolicy, bool, error)

func (fixture policyReaderFixture) GetDeploymentPolicy(
	ctx context.Context,
) (tenancyapi.DeploymentPolicy, bool, error) {
	return fixture(ctx)
}

func TestGetDeploymentPolicyReturnsPersistedAuthority(t *testing.T) {
	zero := int64(0)
	policy, _ := tenancyapi.NewDeploymentPolicy(
		"10000000-0000-4000-8000-000000000001",
		tenancyapi.DeploymentModeOnPremise,
		&zero,
		time.Date(2026, 7, 22, 9, 30, 0, 0, time.UTC),
	)
	query, err := New(policyReaderFixture(
		func(context.Context) (tenancyapi.DeploymentPolicy, bool, error) {
			return policy, true, nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	result, err := query.GetDeploymentPolicy(context.Background())
	if err != nil || result.EnvironmentID() != policy.EnvironmentID() ||
		result.Mode() != tenancyapi.DeploymentModeOnPremise {
		t.Fatalf("policy = %#v, %v", result, err)
	}
}

func TestGetDeploymentPolicyFailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		reader PolicyReader
		want   error
	}{
		{"missing", policyReaderFixture(func(context.Context) (tenancyapi.DeploymentPolicy, bool, error) {
			return tenancyapi.DeploymentPolicy{}, false, nil
		}), tenancyapi.ErrDeploymentPolicyMissing},
		{"invalid", policyReaderFixture(func(context.Context) (tenancyapi.DeploymentPolicy, bool, error) {
			return tenancyapi.DeploymentPolicy{}, true, nil
		}), tenancyapi.ErrDeploymentPolicyUnavailable},
		{"storage", policyReaderFixture(func(context.Context) (tenancyapi.DeploymentPolicy, bool, error) {
			return tenancyapi.DeploymentPolicy{}, false, errors.New("storage failed")
		}), tenancyapi.ErrDeploymentPolicyUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query, err := New(test.reader)
			if err != nil {
				t.Fatal(err)
			}
			_, err = query.GetDeploymentPolicy(context.Background())
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestGetDeploymentPolicyRejectsMissingReaderAndContext(t *testing.T) {
	if _, err := New(nil); !errors.Is(
		err, tenancyapi.ErrDeploymentPolicyUnavailable,
	) {
		t.Fatalf("constructor error = %v", err)
	}
	query, _ := New(policyReaderFixture(
		func(context.Context) (tenancyapi.DeploymentPolicy, bool, error) {
			t.Fatal("reader must not be called")
			return tenancyapi.DeploymentPolicy{}, false, nil
		},
	))
	if _, err := query.GetDeploymentPolicy(nil); !errors.Is(
		err, tenancyapi.ErrDeploymentPolicyUnavailable,
	) {
		t.Fatalf("nil context error = %v", err)
	}
}
