package iam

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"strconv"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ProfileBackfillPublicationRuntime) readPage(
	ctx context.Context,
	scope platformtx.TenantScope,
	source profileapp.Schema,
	evidence profileapp.MutationEvidence,
	after uuid.UUID,
) ([]profilebackfillapp.SourceResource, error) {
	rows, err := runtime.reader.Read(ctx, scope, profilebackfillapp.ReadQuery{
		ResourceKind: string(source.Resource()), SourceSchemaID: source.ID(),
		AfterResourceID: after, Limit: profilebackfillapp.MaximumPageSize,
		Evidence: profileBackfillEvidence(evidence),
	})
	if err != nil {
		return nil, err
	}
	previous := after
	for _, row := range rows {
		if row.Validate() != nil || previous != uuid.Nil &&
			bytes.Compare(previous[:], row.ID[:]) >= 0 {
			wipeProfileSources(rows)
			return nil, profilebackfillapp.ErrUnavailable
		}
		previous = row.ID
	}
	return rows, nil
}

func profileBackfillPage(
	planID uuid.UUID,
	pageNumber int,
	runAt time.Time,
	after uuid.UUID,
	source profileapp.Schema,
	target profileapp.Schema,
	mappings []profileapp.MigrationMapping,
	schemas map[uuid.UUID]profileapp.Schema,
	rows []profilebackfillapp.SourceResource,
) (profilebackfillapp.Page, error) {
	if len(rows) == 0 || len(rows) > profilebackfillapp.MaximumPageSize ||
		len(schemas) == 0 {
		return profilebackfillapp.Page{}, profilebackfillapp.ErrInvalid
	}
	sourceDigest, targetDigest := sha256.New(), sha256.New()
	selected := make(map[uuid.UUID][]profileapp.MigrationMapping, len(schemas))
	for _, row := range rows {
		writeProfileBackfillDigest(
			sourceDigest, []byte(row.ID.String()),
			[]byte(row.SchemaVersionID.String()),
			[]byte(strconv.FormatInt(row.Revision, 10)), row.SHA256[:],
		)
		actual, found := schemas[row.SchemaVersionID]
		if !found || actual.ID() != row.SchemaVersionID ||
			actual.Resource() != source.Resource() {
			return profilebackfillapp.Page{}, profilebackfillapp.ErrUnavailable
		}
		actualMappings, found := selected[row.SchemaVersionID]
		if !found {
			var err error
			actualMappings, err = profileapp.SelectMigrationMappingsForSource(
				actual, target, mappings,
			)
			if err != nil {
				return profilebackfillapp.Page{}, profilebackfillapp.ErrUnavailable
			}
			selected[row.SchemaVersionID] = actualMappings
		}
		document, err := profileBackfillDocument(actual, target, actualMappings, row)
		if err != nil {
			return profilebackfillapp.Page{}, ErrProfileBackfillUnmappable
		}
		digest := document.SHA256()
		writeProfileBackfillDigest(targetDigest, []byte(row.ID.String()), digest[:])
	}
	page := profilebackfillapp.Page{
		JobID:           profileBackfillJobID(planID, pageNumber),
		ScheduledRunAt:  profileBackfillRunAt(runAt, pageNumber),
		AfterResourceID: after, ThroughResourceID: rows[len(rows)-1].ID,
		ExpectedCount: len(rows),
	}
	copy(page.SourceSHA256[:], sourceDigest.Sum(nil))
	copy(page.ExpectedTargetSHA256[:], targetDigest.Sum(nil))
	return page, nil
}

func writeProfileBackfillDigest(digest hash.Hash, values ...[]byte) {
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = digest.Write(size[:])
		_, _ = digest.Write(value)
	}
}

func wipeProfileSources(resources []profilebackfillapp.SourceResource) {
	for index := range resources {
		clear(resources[index].Document)
		resources[index].Document = nil
	}
}
