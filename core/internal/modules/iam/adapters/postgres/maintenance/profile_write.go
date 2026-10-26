package maintenance

import (
	"context"
	"crypto/sha256"
	"fmt"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *store) applyProfileResource(
	ctx context.Context,
	claimValues []any,
	page profilePage,
	resource profileResource,
	document profileapp.Document,
) error {
	canonical := document.CanonicalJSON()
	digest := document.SHA256()
	defer clear(canonical)
	arguments := append(append([]any(nil), claimValues...),
		page.planID, page.pageNumber, resource.id, resource.actualSchemaID,
		resource.revision, resource.sha, canonical, digest[:],
	)
	var revision int64
	if err := store.database.QueryRow(
		ctx, applyProfileBackfillSQL, arguments...,
	).Scan(&revision); err != nil {
		return fmt.Errorf("apply profile backfill resource: %w", err)
	}
	if revision != resource.revision+1 {
		return ErrRuntimeInvalid
	}
	return nil
}

func (store *store) recordProfileFailure(
	ctx context.Context,
	claimValues []any,
	page profilePage,
	resource profileResource,
	code string,
) ([]byte, error) {
	digest := sha256.New()
	writeDirectoryDigest(digest, []byte(resource.id.String()),
		[]byte(resource.actualSchemaID.String()), resource.sha, []byte(code))
	evidence := digest.Sum(nil)
	arguments := append(append([]any(nil), claimValues...),
		page.planID, page.pageNumber, resource.id, code, evidence,
	)
	var recorded bool
	if err := store.database.QueryRow(
		ctx, recordProfileErrorSQL, arguments...,
	).Scan(&recorded); err != nil || !recorded {
		clear(evidence)
		return nil, fmt.Errorf("record profile backfill failure: %w", err)
	}
	return evidence, nil
}

func finishProfileArguments(
	claimValues []any,
	page profilePage,
	scanned int64,
	converted int64,
	resultSHA []byte,
) []any {
	return append(append([]any(nil), claimValues...),
		page.planID, page.pageNumber, scanned, converted,
		page.sourceSHA, resultSHA,
	)
}
