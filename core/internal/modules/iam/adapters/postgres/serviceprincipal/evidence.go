package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
	"github.com/google/uuid"
)

func (bound *boundEvidenceReader) readEvidence(
	ctx context.Context,
	grantID uuid.UUID,
	grantRevision int64,
) (serviceapp.Principal, serviceapp.GrantRevision, error) {
	var document []byte
	if err := bound.database.QueryRow(
		ctx, evidenceSQL, grantID, grantRevision,
	).Scan(&document); err != nil || len(document) == 0 {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	principal, grant, err := decodeEvidence(document)
	clear(document)
	if err != nil {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	return principal, grant, nil
}

func (reader *evidenceReader) Evidence(
	ctx context.Context,
	scope tenantScope,
	grantID uuid.UUID,
	grantRevision int64,
) (serviceapp.Principal, serviceapp.GrantRevision, error) {
	if grantID == uuid.Nil || grantRevision < 1 || !evidenceScopeUsable(ctx, scope) {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	bound, err := reader.bind(scope)
	if err != nil {
		return serviceapp.Principal{}, serviceapp.GrantRevision{}, errEvidenceRead
	}
	return bound.readEvidence(ctx, grantID, grantRevision)
}
