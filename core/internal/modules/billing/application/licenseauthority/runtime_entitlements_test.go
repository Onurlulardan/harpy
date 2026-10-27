package licenseauthority

import "testing"

func TestRuntimeEntitlementsPreserveGraceAndBuildOffOnlyWidening(t *testing.T) {
	fixture := newAuthorityFixture(t)
	snapshot, err := Verify(
		t.Context(), fixture.document(t), fixture.binding, fixture.now, fixture.resolver(),
	)
	if err != nil {
		t.Fatal(err)
	}
	grace := ProjectRuntimeEntitlements(
		snapshot, snapshot.LicenseGrantDeadline().Add(1),
	)
	if grace.State != RuntimeGrace || !grace.Projection.Boolean["forms.builder.enabled"] ||
		!grace.Projection.Limits["forms.active_forms"].Unlimited {
		t.Fatalf("grace narrowed signed grants: %+v", grace)
	}
	buildOff := ProjectRuntimeEntitlements(snapshot, snapshot.GraceUntil().Add(1))
	if buildOff.State != RuntimeBuildOff ||
		buildOff.Projection.Boolean["forms.builder.enabled"] ||
		!buildOff.Projection.Boolean["api.access"] ||
		buildOff.Projection.Limits["forms.active_forms"].Unlimited ||
		buildOff.Projection.Limits["forms.active_forms"].Value != 0 {
		t.Fatalf("build-off safety projection drifted: %+v", buildOff)
	}
}

func TestSignedTerminalStateSkipsGrace(t *testing.T) {
	fixture := newAuthorityFixture(t)
	fixture.payload.Status = StatusRevoked
	snapshot, err := Verify(
		t.Context(), fixture.document(t), fixture.binding, fixture.now, fixture.resolver(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result := ProjectRuntimeEntitlements(snapshot, fixture.now); result.State != RuntimeBuildOff {
		t.Fatalf("revoked snapshot retained grants: %+v", result)
	}
}
