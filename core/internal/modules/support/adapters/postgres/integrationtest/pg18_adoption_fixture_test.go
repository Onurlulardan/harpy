//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"sort"
	"testing"
	"time"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

type adoptionClock struct{ value time.Time }

func (clock *adoptionClock) Now() time.Time { return clock.value }

type adoptionSource struct {
	locator  application.VerifiedLocator
	manifest application.SourceManifest
	contacts []application.ContactSourceRow
	bugs     []application.BugSourceRow
}

func (source *adoptionSource) Inspect(_ context.Context,
	locator application.VerifiedLocator,
) (application.SourceManifest, error) {
	if source == nil || locator != source.locator {
		return application.SourceManifest{}, application.ErrSourceDrift
	}
	return source.manifest, nil
}

func (source *adoptionSource) ReadContacts(_ context.Context,
	locator application.VerifiedLocator, after uuid.UUID, limit int,
) (application.ContactPage, error) {
	if source == nil || locator != source.locator || limit != application.PageLimit {
		return application.ContactPage{}, application.ErrSourceDrift
	}
	rows := contactRowsAfter(source.contacts, after, limit)
	return application.ContactPage{ManifestSHA256: source.manifest.ManifestSHA256,
		After: after, Rows: rows, Done: len(rows) == len(contactRowsAfter(source.contacts, after, 10000))}, nil
}

func (source *adoptionSource) ReadBugs(_ context.Context,
	locator application.VerifiedLocator, after uuid.UUID, limit int,
) (application.BugPage, error) {
	if source == nil || locator != source.locator || limit != application.PageLimit {
		return application.BugPage{}, application.ErrSourceDrift
	}
	rows := bugRowsAfter(source.bugs, after, limit)
	return application.BugPage{ManifestSHA256: source.manifest.ManifestSHA256,
		After: after, Rows: rows, Done: len(rows) == len(bugRowsAfter(source.bugs, after, 10000))}, nil
}

func contactRowsAfter(rows []application.ContactSourceRow, after uuid.UUID,
	limit int,
) []application.ContactSourceRow {
	result := make([]application.ContactSourceRow, 0, limit)
	for _, row := range rows {
		if after != uuid.Nil && row.ID.String() <= after.String() {
			continue
		}
		result = append(result, row)
		if len(result) == limit {
			break
		}
	}
	return result
}

func bugRowsAfter(rows []application.BugSourceRow, after uuid.UUID,
	limit int,
) []application.BugSourceRow {
	result := make([]application.BugSourceRow, 0, limit)
	for _, row := range rows {
		if after != uuid.Nil && row.ID.String() <= after.String() {
			continue
		}
		result = append(result, row)
		if len(result) == limit {
			break
		}
	}
	return result
}

func contactSourceFixture(t *testing.T, locator application.VerifiedLocator,
	rows []application.ContactSourceRow, manifestName string,
) *adoptionSource {
	t.Helper()
	sort.Slice(rows, func(left, right int) bool { return rows[left].ID.String() < rows[right].ID.String() })
	pageSHA, err := application.ContactSourcePageSHA256(rows)
	if err != nil {
		t.Fatal(err)
	}
	root := application.EmptySourceRoot()
	if len(rows) > 0 {
		root = application.FoldSHA256(root, pageSHA)
	}
	return &adoptionSource{locator: locator, contacts: rows,
		manifest: application.SourceManifest{SnapshotID: "snapshot-" + manifestName,
			RowCount: int64(len(rows)), RootSHA256: root,
			ManifestSHA256: sha256.Sum256([]byte(manifestName))}}
}

func bugSourceFixture(t *testing.T, locator application.VerifiedLocator,
	rows []application.BugSourceRow, manifestName string,
) *adoptionSource {
	t.Helper()
	sort.Slice(rows, func(left, right int) bool { return rows[left].ID.String() < rows[right].ID.String() })
	pageSHA, err := application.BugSourcePageSHA256(rows)
	if err != nil {
		t.Fatal(err)
	}
	root := application.EmptySourceRoot()
	if len(rows) > 0 {
		root = application.FoldSHA256(root, pageSHA)
	}
	return &adoptionSource{locator: locator, bugs: rows,
		manifest: application.SourceManifest{SnapshotID: "snapshot-" + manifestName,
			RowCount: int64(len(rows)), RootSHA256: root,
			ManifestSHA256: sha256.Sum256([]byte(manifestName))}}
}
