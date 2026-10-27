package systemtransferartifact

import (
	"context"
	"errors"
	"testing"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

const cleanupObjectKey = "iam-system-import-sources/v1/019bffff-ffff-7fff-8fff-000000000001/1.fcart"

func TestDeleteExactImportSourceReturnsAbsenceEvidence(t *testing.T) {
	runtime, imports, _ := newRuntimeFixture(t)
	result, err := runtime.DeleteExactImportSource(
		context.Background(), cleanupObjectKey, "version-1",
	)
	if err != nil || result.Validate() != nil || imports.deletes != 1 {
		t.Fatalf("DeleteExactImportSource() = %#v, %v; deletes=%d",
			result, err, imports.deletes)
	}
}

func TestDeleteExactImportSourceRejectsUnboundedIdentity(t *testing.T) {
	runtime, imports, _ := newRuntimeFixture(t)
	for _, key := range []string{
		"iam-system-exports/v1/019bffffffff7fff8fff000000000001/1.fcart",
		"iam-system-import-sources/v1/not-a-uuid/1.fcart",
		cleanupObjectKey + "/extra",
	} {
		if _, err := runtime.DeleteExactImportSource(
			context.Background(), key, "version-1",
		); !errors.Is(err, systemtransfer.ErrSourceCleanupInvalid) {
			t.Fatalf("DeleteExactImportSource(%q) error = %v", key, err)
		}
	}
	if imports.deletes != 0 {
		t.Fatalf("invalid identities caused %d deletes", imports.deletes)
	}
}

func TestDeleteExactImportSourceMapsUnknownOutcome(t *testing.T) {
	runtime, imports, _ := newRuntimeFixture(t)
	imports.deleteErr = artifactobject.ErrOutcomeUnknown
	if _, err := runtime.DeleteExactImportSource(
		context.Background(), cleanupObjectKey, "version-1",
	); !errors.Is(err, systemtransfer.ErrSourceDeleteOutcomeUnknown) {
		t.Fatalf("DeleteExactImportSource() error = %v", err)
	}
}
