package systemtransfer

import (
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type outboundAttemptRow struct {
	actorID                   uuid.UUID
	token, authorization      int64
	catalogVersionID          uuid.UUID
	catalogSHA, queryDocument []byte
	querySHA                  []byte
	format, state             string
	attemptGeneration         *int64
}

func (raw *outboundAttemptRow) targets() []any {
	return []any{&raw.actorID, &raw.token, &raw.authorization,
		&raw.catalogVersionID, &raw.catalogSHA, &raw.queryDocument, &raw.querySHA,
		&raw.format, &raw.state, &raw.attemptGeneration}
}

func (raw outboundAttemptRow) decode(claim app.ExportJobClaim) (app.ExportAttempt, error) {
	catalog, catalogOK := workerDigest(raw.catalogSHA)
	query, queryOK := workerDigest(raw.querySHA)
	generation := int64(0)
	if raw.attemptGeneration != nil {
		generation = *raw.attemptGeneration
	}
	attempt := app.ExportAttempt{TenantID: claim.TenantID, ActorID: raw.actorID,
		JobID: claim.JobID, ExportID: claim.ExportID,
		CatalogVersionID: raw.catalogVersionID, Resource: claim.Resource,
		Format: raw.format, State: app.ExportOwnerState(raw.state),
		TokenVersion: raw.token, AuthorizationVersion: raw.authorization,
		CatalogSHA256: catalog, QuerySHA256: query,
		QueryDocument:     append([]byte(nil), raw.queryDocument...),
		AttemptGeneration: generation}
	if !catalogOK || !queryOK || attempt.CheckOwner() != nil {
		return app.ExportAttempt{}, app.ErrExportStateConflict
	}
	return attempt, nil
}
