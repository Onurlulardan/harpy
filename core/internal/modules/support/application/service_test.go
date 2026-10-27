package application

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func TestHoneypotContactIsAcknowledgedWithoutPersistence(t *testing.T) {
	store := &storeFixture{}
	service := serviceFixture(store)
	err := service.ContactSubmit(context.Background(), domain.ContactInput{
		Honeypot: "bot-filled", Name: "", Email: ""}, [32]byte{}, "", [32]byte{})
	if err != nil || store.contactSubmissions != 0 {
		t.Fatalf("honeypot result err=%v submissions=%d", err, store.contactSubmissions)
	}
}

func TestReporterReadPinsTenantAndRedactsAdminNotes(t *testing.T) {
	access := accessFixture(ScopeReporter, PermissionBugReadOwn, OperationBugRead)
	store := &storeFixture{bug: domain.BugReport{ID: domain.ID(uuid.MustParse(
		"01900000-0000-7000-8000-000000000002")), TenantID: domain.ID(access.TenantID()),
		ReporterID: domain.ID(access.ActorID()), AdminNotes: "secret operator note"}}
	service := serviceFixture(store)
	report, err := service.ReporterBugGet(context.Background(), access, store.bug.ID)
	if err != nil {
		t.Fatal(err)
	}
	if store.lastTenant != access.TenantID() || report.AdminNotes != "" {
		t.Fatalf("reporter boundary failed tenant=%s admin_notes=%q",
			store.lastTenant, report.AdminNotes)
	}
}

func TestTenantReadRejectsReporterCapability(t *testing.T) {
	access := accessFixture(ScopeReporter, PermissionBugReadOwn, OperationBugRead)
	service := serviceFixture(&storeFixture{})
	_, err := service.TenantBugGet(context.Background(), access,
		domain.ID(uuid.MustParse("01900000-0000-7000-8000-000000000002")))
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("cross-scope read error = %v, want forbidden", err)
	}
}

func TestContactSubmissionRequiresDurableCommandEvidence(t *testing.T) {
	service := serviceFixture(&storeFixture{})
	input := domain.ContactInput{Name: "Ada", Email: "ada@example.com",
		Message: "Need help", ContactType: "general"}
	rate := sha256.Sum256([]byte("rate"))
	_, empty := sha256.Sum256([]byte("request")), [32]byte{}
	if err := service.ContactSubmit(context.Background(), input, rate,
		"contact-command-0001", empty); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing digest error = %v, want invalid", err)
	}
}
