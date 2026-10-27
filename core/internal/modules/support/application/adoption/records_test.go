package adoption

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestContactSourceEvidenceIsDeterministicAndNormalizesOnlyTarget(t *testing.T) {
	now := time.Date(2026, 8, 1, 10, 0, 0, 123456000, time.UTC)
	row := ContactSourceRow{ID: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		Name: " Ada ", Email: " ADA@EXAMPLE.COM ", Message: " Need help ",
		ContactType: "general", Status: "new", IPAddress: "192.0.2.1",
		UserAgent: "legacy", CreatedAt: now}
	first, err := contactRecord(row)
	if err != nil {
		t.Fatal(err)
	}
	second, err := contactRecord(row)
	if err != nil || first.SourceSHA256 != second.SourceSHA256 || string(first.Target) != string(second.Target) {
		t.Fatal("contact transform is not deterministic")
	}
	var target map[string]any
	if json.Unmarshal(first.Target, &target) != nil || target["name"] != "Ada" ||
		target["email"] != "ada@example.com" {
		t.Fatalf("unexpected normalized target %s", first.Target)
	}
	changed := row
	changed.IPAddress = "192.0.2.2"
	changedRecord, err := contactRecord(changed)
	if err != nil || changedRecord.SourceSHA256 == first.SourceSHA256 ||
		string(changedRecord.Target) != string(first.Target) {
		t.Fatal("redacted legacy evidence was not isolated from the target transform")
	}
}

func TestSourcePagesRejectOrderDeletionAndUnstableShortPages(t *testing.T) {
	now := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	manifest := sha256.Sum256([]byte("manifest"))
	run := Run{ManifestSHA256: manifest}
	high := ContactSourceRow{ID: uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		Name: "Ada", Email: "ada@example.com", Message: "Need help",
		ContactType: "general", Status: "new", CreatedAt: now}
	low := high
	low.ID = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	_, err := contactRecords(run, ContactPage{ManifestSHA256: manifest,
		Rows: []ContactSourceRow{high, low}, Done: true})
	if !errors.Is(err, ErrSourceDrift) {
		t.Fatalf("unordered page error=%v", err)
	}
	_, err = contactRecords(run, ContactPage{ManifestSHA256: manifest,
		Rows: []ContactSourceRow{low}, Done: false})
	if !errors.Is(err, ErrSourceDrift) {
		t.Fatalf("short non-final page error=%v", err)
	}
	deletedAt := now
	bug := BugSourceRow{ID: low.ID, TenantID: uuid.New(), ReporterID: uuid.New(),
		ReporterName: "Ada", ReporterEmail: "ada@example.com", Title: "Legacy bug",
		Description: "Legacy description", Category: "bug", Severity: "high",
		Priority: "normal", BrowserInfo: []byte(`{}`), ConsoleErrors: []byte(`[]`),
		NetworkInfo: []byte(`[]`), Screenshots: []byte(`[]`), Status: "open",
		DeletedAt: &deletedAt, CreatedAt: now, UpdatedAt: now}
	if _, bugErr := BugSourcePageSHA256([]BugSourceRow{bug}); !errors.Is(bugErr, ErrSourceDrift) {
		t.Fatalf("deleted source row error=%v", bugErr)
	}
}
