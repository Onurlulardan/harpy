package dashboard

import (
	"crypto/subtle"
	"time"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	"github.com/google/uuid"
)

type snapshotRow struct {
	dashboardID, versionID                       uuid.UUID
	versionNumber, rootRevision, versionRevision int64
	lifecycle, versionState, name, visibility    string
	layout, layoutSHA256, manifest, widgets      []byte
	createdAt, updatedAt                         time.Time
}

func scanSnapshot(scanner row, notFound rowNotFound) (dashboardapp.Snapshot, error) {
	stored := snapshotRow{}
	err := scanner.Scan(&stored.dashboardID, &stored.versionID, &stored.versionNumber,
		&stored.rootRevision, &stored.versionRevision, &stored.lifecycle, &stored.versionState,
		&stored.name, &stored.visibility, &stored.layout, &stored.layoutSHA256,
		&stored.manifest, &stored.widgets, &stored.createdAt, &stored.updatedAt)
	if err != nil {
		return dashboardapp.Snapshot{}, mapError(err, notFound)
	}
	return stored.decode()
}

func scanMutation(scanner row, notFound rowNotFound) (dashboardapp.Snapshot, bool, error) {
	stored := snapshotRow{}
	var replay bool
	err := scanner.Scan(&stored.dashboardID, &stored.versionID, &stored.versionNumber,
		&stored.rootRevision, &stored.versionRevision, &stored.lifecycle, &stored.versionState,
		&stored.name, &stored.visibility, &stored.layout, &stored.layoutSHA256,
		&stored.manifest, &stored.widgets, &stored.createdAt, &stored.updatedAt, &replay)
	if err != nil {
		return dashboardapp.Snapshot{}, false, mapError(err, notFound)
	}
	snapshot, err := stored.decode()
	return snapshot, replay, err
}

func (stored snapshotRow) decode() (dashboardapp.Snapshot, error) {
	dashboardID, dashboardErr := analyticsdashboard.NewID([16]byte(stored.dashboardID))
	versionID, versionErr := analyticsdashboard.NewVersionID([16]byte(stored.versionID))
	layout, layoutErr := dashboardapp.NewDocument(stored.layout)
	widgets, widgetsErr := decodeWidgets(stored.widgets)
	digest := layout.SHA256()
	if dashboardErr != nil || versionErr != nil || layoutErr != nil || widgetsErr != nil ||
		stored.versionNumber < 1 || stored.rootRevision < 1 || stored.versionRevision < 1 ||
		len(stored.layoutSHA256) != 32 || subtle.ConstantTimeCompare(stored.layoutSHA256, digest[:]) != 1 ||
		len(stored.manifest) != 32 || stored.createdAt.IsZero() || stored.updatedAt.Before(stored.createdAt) {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	lifecycle := analyticsdashboard.Lifecycle(stored.lifecycle)
	visibility := analyticsdashboard.Visibility(stored.visibility)
	published := stored.versionState == "published"
	if !lifecycle.Valid() || !visibility.Valid() ||
		(stored.versionState != "draft" && !published && stored.versionState != "retired") {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	var layoutDigest, manifest [32]byte
	copy(layoutDigest[:], stored.layoutSHA256)
	copy(manifest[:], stored.manifest)
	_ = dashboardID
	_ = versionID
	return dashboardapp.Snapshot{DashboardID: stored.dashboardID, VersionID: stored.versionID,
		VersionNumber: uint64(stored.versionNumber), RootRevision: uint64(stored.rootRevision),
		VersionRevision: uint64(stored.versionRevision), Lifecycle: lifecycle,
		Visibility: visibility, Published: published, Name: stored.name,
		LayoutDocument: layout.CanonicalJSON(), LayoutSHA256: layoutDigest,
		ManifestSHA256: manifest, Widgets: widgets,
		CreatedAt: stored.createdAt.UTC(), UpdatedAt: stored.updatedAt.UTC()}, nil
}
