package systemtransferartifact

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestExportArtifactReconciliationMapsExactSealedEvidence(t *testing.T) {
	runtime, _, exports := newRuntimeFixture(t)
	effect := exportReconciliationEffectFixture(t)
	modified := fixtureArtifactTime.Add(-time.Second)
	exports.reconcile = artifactobject.ReconcileResult{
		State:             artifactobject.ReconciliationSealed,
		Identity:          artifactobject.Identity{Key: effect.ObjectKey, Version: "version-1"},
		ProviderRequestID: "head-request-1", EvidenceSHA256: [32]byte{7},
		ObjectModifiedAt: modified, VerifiedAt: fixtureArtifactTime}
	result, err := runtime.ReconcileExportArtifact(t.Context(), effect)
	if err != nil || result.State != systemtransfer.ExportArtifactSealed ||
		result.Artifact.ObjectVersion != "version-1" ||
		result.Artifact.KeyID != effect.Draft.KeyID ||
		result.Artifact.ResultSHA256 == [32]byte{} ||
		result.ProviderRequestID != "head-request-1" ||
		result.EvidenceSHA256 != exports.reconcile.EvidenceSHA256 ||
		exports.reconcileExpected.CiphertextSHA256 != effect.Draft.CiphertextSHA256 {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestExportArtifactReconciliationMapsMissingAndUnknown(t *testing.T) {
	for _, test := range []struct {
		name  string
		state artifactobject.ReconciliationState
		err   error
		want  systemtransfer.ExportArtifactReconciliationState
	}{{"missing", artifactobject.ReconciliationMissing, nil,
		systemtransfer.ExportArtifactMissing},
		{"unknown", artifactobject.ReconciliationUnknown,
			artifactobject.ErrOutcomeUnknown, systemtransfer.ExportArtifactUnknown}} {
		t.Run(test.name, func(t *testing.T) {
			runtime, _, exports := newRuntimeFixture(t)
			effect := exportReconciliationEffectFixture(t)
			exports.reconcile = artifactobject.ReconcileResult{State: test.state,
				ProviderRequestID: "head-request-2", EvidenceSHA256: [32]byte{8},
				VerifiedAt: fixtureArtifactTime}
			exports.reconcileErr = test.err
			result, err := runtime.ReconcileExportArtifact(t.Context(), effect)
			if result.State != test.want || (test.err == nil) != (err == nil) {
				t.Fatalf("result = %#v, err = %v", result, err)
			}
		})
	}
}

func TestExportArtifactReconciliationRejectsNonDeterministicKey(t *testing.T) {
	runtime, _, exports := newRuntimeFixture(t)
	effect := exportReconciliationEffectFixture(t)
	effect.ObjectKey = "iam-system-exports/v1/" + fixtureV7(t).String() + "/1.fcart"
	effect.RequestSHA256 = exportUploadRequestSHA(effect.ObjectKey,
		effect.Draft.CiphertextSHA256, effect.Draft.CiphertextBytes)
	_, err := runtime.ReconcileExportArtifact(t.Context(), effect)
	if !errors.Is(err, errInvalid) || exports.reconcileExpected.Key != "" {
		t.Fatalf("error = %v, expected = %#v", err, exports.reconcileExpected)
	}
}

func exportReconciliationEffectFixture(t *testing.T) systemtransfer.ExportArtifactEffectStart {
	query := []byte(`{"select":["id"],"sort":[{"direction":"asc","field":"id"}]}`)
	exportID, effectID := fixtureV7(t), fixtureV7(t)
	draft := systemtransfer.ExportArtifactDraftEvidence{KeyID: "iam-export-kek-1",
		HeaderSHA256: [32]byte{1}, CiphertextSHA256: [32]byte{2},
		CiphertextBytes: 256, WrappedDEK: make([]byte, 32), WrapNonce: [12]byte{1}}
	key := "iam-system-exports/v1/" + exportID.String() + "/1.fcart"
	return systemtransfer.ExportArtifactEffectStart{Attempt: systemtransfer.ExportAttempt{
		TenantID: fixtureV7(t), ActorID: fixtureV7(t), JobID: fixtureV7(t),
		ExportID: exportID, AttemptID: effectID, LeaseOwnerID: fixtureV7(t),
		CatalogVersionID: fixtureV7(t), Resource: publicresource.Users,
		Format: "jsonl", State: systemtransfer.ExportOutcomeUnknown,
		TokenVersion: 1, AuthorizationVersion: 1,
		CatalogSHA256: [32]byte{3}, QuerySHA256: sha256.Sum256(query),
		QueryDocument: query, AttemptGeneration: 1},
		EffectID: effectID, EffectReceiptID: fixtureV7(t), ObjectKey: key,
		RequestSHA256: exportUploadRequestSHA(key, draft.CiphertextSHA256,
			draft.CiphertextBytes),
		StartedAt: fixtureArtifactTime.Add(-2 * time.Second), Draft: draft}
}
