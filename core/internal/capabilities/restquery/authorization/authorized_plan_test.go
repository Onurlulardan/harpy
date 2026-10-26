package authorization

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestBuildAuthorizedPlanIsCanonicalImmutableAndRedacted(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	plan, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, evaluations,
	)
	if err != nil || plan.Validate() != nil {
		t.Fatalf("build authorized plan: %v", err)
	}
	tenant, tenantErr := plan.TenantID()
	actor, actorErr := plan.ActorID()
	contextHash, contextErr := plan.AuthorizationContextSHA256()
	schema, schemaErr := plan.PolicySchema()
	visibilities, visibilityErr := plan.Visibilities()
	if tenantErr != nil || actorErr != nil || contextErr != nil ||
		schemaErr != nil || visibilityErr != nil || tenant != authorizedPlanTenant ||
		actor != authorizedPlanActor || contextHash == [sha256.Size]byte{} ||
		schema != resolved.AuthorizationPolicySchemaV1 || len(visibilities) != 2 {
		t.Fatal("authorized plan accessors lost sealed state")
	}
	canonical, err := plan.MandatoryCanonical()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(canonical)
	const want = "58ab860c1483c04cdf3dc3705029c405c93f36b5fd0b2ea9a844b379a72e50e6"
	if hex.EncodeToString(digest[:]) != want {
		t.Fatalf("mandatory canonical hash = %s", hex.EncodeToString(digest[:]))
	}
	canonical[0] ^= 0xff
	visibilities[0] = resolved.ResourceVisibility{}
	if plan.Validate() != nil {
		t.Fatal("returned state mutated authorized plan")
	}
	if fmt.Sprint(plan) != authorizationRedaction {
		t.Fatal("authorized plan formatting leaked")
	}
	if _, err := json.Marshal(plan); err == nil {
		t.Fatal("authorized plan allowed direct JSON marshal")
	}
}

func TestBuildAuthorizedPlanRejectsZeroAndOutOfOrderEvaluation(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	if _, err := buildAuthorizedPlan(
		resolved.QuerySpec{}, set, authorizedPlanTenant,
		authorizedPlanActor, evaluations,
	); err == nil {
		t.Fatal("zero QuerySpec was authorized")
	}
	evaluations[0], evaluations[1] = evaluations[1], evaluations[0]
	if _, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, evaluations,
	); err == nil {
		t.Fatal("out-of-order owner evaluations were authorized")
	}
}
