package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestProjectionPersistenceRequiresCanonicalVisibilityPrincipals(t *testing.T) {
	id := domain.ID{1}
	valid := ports.ProjectionItem{ID: id, SourceID: id, SourceVersionID: id,
		SourceOwner: domain.SourceWorkflow, SourceKind: "HUMAN_TASK", Surface: "INBOX",
		Summary: "approval", SourceRevision: 1, ReferenceEpoch: 1,
		PermissionRevision: 1, EventSequence: 1,
		VisibleTo: []domain.Principal{{Kind: domain.PrincipalUser, ID: id}}}
	if !validProjectionItem(valid) {
		t.Fatal("valid projection visibility was rejected")
	}
	valid.VisibleTo = nil
	if validProjectionItem(valid) {
		t.Fatal("projection without visibility principals was accepted")
	}
}
