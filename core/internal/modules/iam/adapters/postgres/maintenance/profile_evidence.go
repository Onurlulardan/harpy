package maintenance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strconv"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) verifyProfilePageEvidence(
	ctx context.Context,
	claim maintenanceapp.Claim,
	page profilePage,
) error {
	values, err := claimValues(claim)
	if err != nil {
		return err
	}
	rows, err := store.database.Query(
		ctx, readProfileResourceEvidenceSQL,
		append(values, page.planID, page.pageNumber)...,
	)
	if err != nil {
		return fmt.Errorf("read profile page evidence: %w", err)
	}
	defer rows.Close()
	digest := sha256.New()
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var actualSchemaID uuid.UUID
		var revision int64
		var sourceSHA []byte
		if err := rows.Scan(&id, &actualSchemaID, &revision, &sourceSHA); err != nil ||
			id == uuid.Nil || actualSchemaID == uuid.Nil ||
			revision < 1 || len(sourceSHA) != sha256.Size {
			clear(sourceSHA)
			return ErrRuntimeInvalid
		}
		writeDirectoryDigest(digest, []byte(id.String()), []byte(actualSchemaID.String()),
			[]byte(strconv.FormatInt(revision, 10)), sourceSHA)
		clear(sourceSHA)
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate profile page evidence: %w", err)
	}
	result := digest.Sum(nil)
	defer clear(result)
	if count != page.expectedCount || !hmac.Equal(result, page.sourceSHA) {
		return ErrRuntimeInvalid
	}
	return nil
}
