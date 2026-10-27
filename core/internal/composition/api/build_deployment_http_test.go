package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tenancymodule "github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
)

type deploymentPolicyQueryStub struct {
	policy tenancymodule.DeploymentPolicy
}

func (stub deploymentPolicyQueryStub) GetDeploymentPolicy(
	context.Context,
) (tenancymodule.DeploymentPolicy, error) {
	return stub.policy, nil
}

type primaryTargetQueryStub struct {
	target tenancymodule.PrimaryTenantTarget
}

func (stub primaryTargetQueryStub) GetPrimaryTenantTarget(
	context.Context,
) (tenancymodule.PrimaryTenantTarget, error) {
	return stub.target, nil
}

func TestDeploymentHTTPUsesVerifiedPersistedAuthority(t *testing.T) {
	t.Parallel()
	runtime := verifiedDeploymentRuntime(t, tenancymodule.DeploymentModeOnPremise)
	foundation, err := buildDeploymentHTTPFoundation(
		runtimeConfigInfo(t), runtime, func(error) {},
	)
	if err != nil || len(foundation.routes) != 2 {
		t.Fatalf("deployment HTTP = %#v, %v", foundation, err)
	}
	recorder := httptest.NewRecorder()
	foundation.routes[0].Handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/app/config", nil),
	)
	var response runtimeConfigEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || response.Data.DeploymentMode != "onpremise" ||
		response.Data.DefaultTenantSlug == nil ||
		*response.Data.DefaultTenantSlug != "harpy" ||
		!response.Data.ServiceAvailability.Work {
		t.Fatalf("persisted runtime response = %#v", response)
	}
}

func TestDeploymentHTTPRejectsUnverifiedRuntime(t *testing.T) {
	t.Parallel()
	runtime, err := tenancymodule.NewDeploymentPolicyRuntime(
		string(tenancymodule.DeploymentModeSaaS),
		deploymentPolicyQueryStub{policy: billingPolicyFixture(
			t, tenancymodule.DeploymentModeSaaS,
		)},
		primaryTargetQueryStub{target: primaryTargetFixture(
			t, tenancymodule.DeploymentModeSaaS,
		)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := buildDeploymentHTTPFoundation(
		runtimeConfigInfo(t), runtime, func(error) {},
	); err == nil {
		t.Fatal("unverified deployment runtime exposed public config")
	}
}

func verifiedDeploymentRuntime(
	t *testing.T,
	mode tenancymodule.DeploymentMode,
) *tenancymodule.DeploymentPolicyRuntime {
	t.Helper()
	runtime, err := tenancymodule.NewDeploymentPolicyRuntime(
		string(mode),
		deploymentPolicyQueryStub{policy: billingPolicyFixture(t, mode)},
		primaryTargetQueryStub{target: primaryTargetFixture(t, mode)},
	)
	if err != nil || runtime.Verify(context.Background()) != nil {
		t.Fatalf("verify deployment runtime = %v, %v", runtime, err)
	}
	return runtime
}

func primaryTargetFixture(
	t *testing.T,
	mode tenancymodule.DeploymentMode,
) tenancymodule.PrimaryTenantTarget {
	t.Helper()
	purpose := tenancymodule.PrimaryTenantPurposeOnPremise
	slug := "harpy"
	if mode == tenancymodule.DeploymentModeSaaS {
		purpose = tenancymodule.PrimaryTenantPurposeSaaS
		slug = "control"
	}
	target, err := tenancymodule.NewPrimaryTenantTarget(
		"10000000-0000-4000-8000-000000000001", mode,
		"20000000-0000-4000-8000-000000000001", purpose, slug,
	)
	if err != nil {
		t.Fatal(err)
	}
	return target
}
