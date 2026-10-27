package deploymentpolicy

import (
	"context"
	"errors"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
)

func TestCreateDeploymentPolicyReturnsExactOwnerReceipt(t *testing.T) {
	request := initializationRequest()
	created := time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)
	maximum := int64(0)
	policy, err := tenantapi.NewDeploymentPolicy(
		tenantapi.UUIDString(request.EnvironmentID), request.DeploymentMode,
		&maximum, created)
	if err != nil {
		t.Fatal(err)
	}
	store := initializationStoreFunc(func(
		context.Context, tenantapi.InitializeDeploymentPolicyRequest,
	) (tenantapi.DeploymentPolicyInitialization, error) {
		return tenantapi.DeploymentPolicyInitialization{OperationID: request.OperationID,
			Policy: policy, DatabaseNow: created, Replayed: true}, nil
	})
	initializer, _ := NewInitializer(store)
	receipt, err := initializer.CreateDeploymentPolicy(context.Background(), request)
	if err != nil || !receipt.IsValid() || !receipt.Replayed {
		t.Fatalf("receipt = %#v, %v", receipt, err)
	}
}

func TestCreateDeploymentPolicyFailsClosed(t *testing.T) {
	initializer, _ := NewInitializer(initializationStoreFunc(func(
		context.Context, tenantapi.InitializeDeploymentPolicyRequest,
	) (tenantapi.DeploymentPolicyInitialization, error) {
		return tenantapi.DeploymentPolicyInitialization{},
			tenantapi.ErrDeploymentPolicyConflict
	}))
	if _, err := initializer.CreateDeploymentPolicy(
		context.Background(), initializationRequest(),
	); !errors.Is(err, tenantapi.ErrDeploymentPolicyConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if _, err := initializer.CreateDeploymentPolicy(context.Background(),
		tenantapi.InitializeDeploymentPolicyRequest{}); !errors.Is(
		err, tenantapi.ErrDeploymentPolicyInvalid) {
		t.Fatalf("invalid error = %v", err)
	}
}

type initializationStoreFunc func(
	context.Context,
	tenantapi.InitializeDeploymentPolicyRequest,
) (tenantapi.DeploymentPolicyInitialization, error)

func (function initializationStoreFunc) Create(
	ctx context.Context,
	request tenantapi.InitializeDeploymentPolicyRequest,
) (tenantapi.DeploymentPolicyInitialization, error) {
	return function(ctx, request)
}

func initializationRequest() tenantapi.InitializeDeploymentPolicyRequest {
	return tenantapi.InitializeDeploymentPolicyRequest{
		OperationID:    [16]byte{0x01, 0x8f, 0, 0, 0, 0, 0x70, 0, 0x80, 0, 0, 0, 0, 0, 0, 1},
		EnvironmentID:  [16]byte{0x01, 0x8f, 0, 0, 0, 0, 0x70, 0, 0x80, 0, 0, 0, 0, 0, 0, 2},
		OperationKind:  tenantapi.OnboardingEnvironmentInitialize,
		DeploymentMode: tenantapi.DeploymentModeOnPremise,
	}
}
