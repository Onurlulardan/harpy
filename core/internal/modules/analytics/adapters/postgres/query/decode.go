package query

import (
	"crypto/subtle"
	"time"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	"github.com/google/uuid"
)

type snapshotRow struct {
	queryID, versionID                                    uuid.UUID
	versionNumber, rootRevision, referenceEpoch, revision int64
	lifecycle, versionState, name                         string
	queryDocument, parameters, output, manifest           []byte
	cachePolicy                                           string
	createdAt, updatedAt                                  time.Time
}

func scanSnapshot(scanner row, notFound rowNotFound) (queryapp.Snapshot, error) {
	stored := snapshotRow{}
	err := scanner.Scan(&stored.queryID, &stored.versionID, &stored.versionNumber,
		&stored.rootRevision, &stored.referenceEpoch, &stored.revision,
		&stored.lifecycle, &stored.versionState, &stored.name, &stored.queryDocument,
		&stored.parameters, &stored.output, &stored.manifest, &stored.cachePolicy,
		&stored.createdAt, &stored.updatedAt)
	if err != nil {
		return queryapp.Snapshot{}, mapError(err, notFound)
	}
	return stored.decode()
}

func scanMutation(scanner row, notFound rowNotFound) (queryapp.Snapshot, bool, error) {
	stored := snapshotRow{}
	var replay bool
	err := scanner.Scan(&stored.queryID, &stored.versionID, &stored.versionNumber,
		&stored.rootRevision, &stored.referenceEpoch, &stored.revision,
		&stored.lifecycle, &stored.versionState, &stored.name, &stored.queryDocument,
		&stored.parameters, &stored.output, &stored.manifest, &stored.cachePolicy,
		&stored.createdAt, &stored.updatedAt, &replay)
	if err != nil {
		return queryapp.Snapshot{}, false, mapError(err, notFound)
	}
	snapshot, err := stored.decode()
	return snapshot, replay, err
}

func (stored snapshotRow) decode() (queryapp.Snapshot, error) {
	queryID, queryErr := analyticsquery.NewID([16]byte(stored.queryID))
	versionID, versionErr := analyticsquery.NewVersionID([16]byte(stored.versionID))
	queryDocument, queryDocumentErr := queryapp.NewDocument(stored.queryDocument)
	parameters, parametersErr := queryapp.NewDocument(stored.parameters)
	output, outputErr := queryapp.NewDocument(stored.output)
	if queryErr != nil || versionErr != nil || queryDocumentErr != nil ||
		parametersErr != nil || outputErr != nil || len(stored.manifest) != 32 ||
		stored.versionNumber < 1 || stored.rootRevision < 1 || stored.referenceEpoch < 1 ||
		stored.revision < 1 || stored.createdAt.IsZero() || stored.updatedAt.Before(stored.createdAt) {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	lifecycle := analyticsquery.Lifecycle(stored.lifecycle)
	state := analyticsquery.VersionState(stored.versionState)
	policy := analyticspublication.CachePolicy(stored.cachePolicy)
	var manifest [32]byte
	copy(manifest[:], stored.manifest)
	if _, err := analyticsquery.Rehydrate(queryID, lifecycle, uint64(stored.rootRevision),
		uint64(stored.referenceEpoch), versionID); err != nil || !policy.Valid() {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	if _, err := analyticsquery.RehydrateVersion(versionID, queryID, uint64(stored.versionNumber),
		uint64(stored.revision), state, queryDocument.SHA256(), manifest); err != nil {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	if subtle.ConstantTimeCompare(parameters.CanonicalJSON(), stored.parameters) != 1 ||
		subtle.ConstantTimeCompare(output.CanonicalJSON(), stored.output) != 1 {
		return queryapp.Snapshot{}, queryapp.ErrUnavailable
	}
	return queryapp.Snapshot{QueryID: stored.queryID, VersionID: stored.versionID,
		VersionNumber: uint64(stored.versionNumber), RootRevision: uint64(stored.rootRevision),
		ReferenceEpoch: uint64(stored.referenceEpoch), VersionRevision: uint64(stored.revision),
		Lifecycle: lifecycle, VersionState: state, CachePolicy: policy, Name: stored.name,
		QueryDocument: queryDocument.CanonicalJSON(), ParameterDocument: parameters.CanonicalJSON(),
		OutputDocument: output.CanonicalJSON(), ManifestSHA256: manifest,
		CreatedAt: stored.createdAt.UTC(), UpdatedAt: stored.updatedAt.UTC()}, nil
}
