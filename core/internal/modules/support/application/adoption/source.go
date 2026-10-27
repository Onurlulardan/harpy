package adoption

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

const PageLimit = 100

type SourceManifest struct {
	SnapshotID     string
	RowCount       int64
	RootSHA256     [32]byte
	ManifestSHA256 [32]byte
}

func (manifest SourceManifest) valid() bool {
	return boundedText(manifest.SnapshotID, 16, 200) && manifest.RowCount >= 0 &&
		!zeroSHA(manifest.RootSHA256) && !zeroSHA(manifest.ManifestSHA256)
}

type ContactSourceRow struct {
	ID                                                  uuid.UUID
	Name, Email, Company, Phone, EmployeeCount, Message string
	ContactType, Status, Source, IPAddress, UserAgent   string
	Notes                                               string
	RepliedAt                                           *time.Time
	RepliedBy                                           *uuid.UUID
	CreatedAt                                           time.Time
}

type BugSourceRow struct {
	ID, TenantID, ReporterID                             uuid.UUID
	ReporterName, ReporterEmail, Title, Description      string
	Category, Severity, Priority, PageURL, UserAgent     string
	StepsToReproduce, Status, Resolution, AdminNotes     string
	BrowserInfo, ConsoleErrors, NetworkInfo, Screenshots json.RawMessage
	ResolvedAt, ClosedAt, DeletedAt                      *time.Time
	CreatedAt, UpdatedAt                                 time.Time
}

type ContactPage struct {
	ManifestSHA256 [32]byte
	After          uuid.UUID
	Rows           []ContactSourceRow
	Done           bool
}

type BugPage struct {
	ManifestSHA256 [32]byte
	After          uuid.UUID
	Rows           []BugSourceRow
	Done           bool
}

type Source interface {
	Inspect(context.Context, VerifiedLocator) (SourceManifest, error)
	ReadContacts(context.Context, VerifiedLocator, uuid.UUID, int) (ContactPage, error)
	ReadBugs(context.Context, VerifiedLocator, uuid.UUID, int) (BugPage, error)
}

func validContactStatus(value string) bool {
	return domain.ValidContactStatus(domain.ContactStatus(value))
}

func validBugState(row BugSourceRow) bool {
	return domain.ValidBugStatus(domain.BugStatus(row.Status)) &&
		domain.ValidBugPriority(domain.BugPriority(row.Priority))
}
