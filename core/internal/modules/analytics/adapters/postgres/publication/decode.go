package publication

import (
	"time"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/google/uuid"
)

type snapshotRow struct {
	releaseID, resourceVersionID               uuid.UUID
	resourceKind, audience, state, cachePolicy string
	manifest                                   []byte
	requesterID                                uuid.UUID
	approverID                                 dbUUID
	revision, generation                       int64
	publicID                                   []byte
	expiresAt                                  dbInstant
	createdAt, updatedAt                       time.Time
}

func scanSnapshot(scanner row, notFound rowNotFound) (publicationapp.Snapshot, error) {
	stored := snapshotRow{}
	err := scanFields(scanner, &stored, nil)
	if err != nil {
		return publicationapp.Snapshot{}, mapError(err, notFound)
	}
	return stored.decode()
}

func scanMutation(scanner row, notFound rowNotFound) (publicationapp.Snapshot, bool, error) {
	stored := snapshotRow{}
	var replay bool
	err := scanFields(scanner, &stored, &replay)
	if err != nil {
		return publicationapp.Snapshot{}, false, mapError(err, notFound)
	}
	snapshot, err := stored.decode()
	return snapshot, replay, err
}

func scanFields(scanner row, stored *snapshotRow, replay *bool) error {
	fields := []any{&stored.releaseID, &stored.resourceKind, &stored.resourceVersionID,
		&stored.audience, &stored.state, &stored.cachePolicy, &stored.manifest,
		&stored.requesterID, &stored.approverID, &stored.revision, &stored.generation,
		&stored.publicID, &stored.expiresAt, &stored.createdAt, &stored.updatedAt}
	if replay != nil {
		fields = append(fields, replay)
	}
	return scanner.Scan(fields...)
}

func (stored snapshotRow) decode() (publicationapp.Snapshot, error) {
	id, idErr := analyticspublication.NewID([16]byte(stored.releaseID))
	resourceKind := publicationapp.ResourceKind(stored.resourceKind)
	audience := analyticspublication.Audience(stored.audience)
	state := analyticspublication.State(stored.state)
	cachePolicy := analyticspublication.CachePolicy(stored.cachePolicy)
	if idErr != nil || !resourceKind.Valid() || !validUUIDv7(stored.resourceVersionID) ||
		!audience.Valid() || !state.Valid() || !cachePolicy.Valid() || len(stored.manifest) != 32 ||
		stored.requesterID == uuid.Nil || stored.revision < 1 || stored.generation < 1 ||
		stored.createdAt.IsZero() || stored.updatedAt.Before(stored.createdAt) || len(stored.publicID) > 0 && len(stored.publicID) != 16 {
		return publicationapp.Snapshot{}, publicationapp.ErrUnavailable
	}
	_ = id
	var manifest [32]byte
	copy(manifest[:], stored.manifest)
	var publicID [16]byte
	copy(publicID[:], stored.publicID)
	approver := uuid.Nil
	if stored.approverID.Valid {
		approver = stored.approverID.UUID
	}
	var expires *time.Time
	if stored.expiresAt.Valid {
		value := stored.expiresAt.Time.UTC()
		expires = &value
	}
	return publicationapp.Snapshot{ReleaseID: stored.releaseID, ResourceKind: resourceKind,
		ResourceVersionID: stored.resourceVersionID, Audience: audience, State: state,
		CachePolicy: cachePolicy, ManifestSHA256: manifest, RequesterID: stored.requesterID,
		ApproverID: approver, Revision: uint64(stored.revision), Generation: uint64(stored.generation),
		PublicID: publicID, CapabilityExpiresAt: expires,
		CreatedAt: stored.createdAt.UTC(), UpdatedAt: stored.updatedAt.UTC()}, nil
}
