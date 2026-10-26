package ports

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type ContactSubmission struct {
	ID             domain.ID
	Input          domain.ContactInput
	RateKeySHA256  [32]byte
	IdempotencyKey string
	RequestSHA256  [32]byte
	ObservedAt     time.Time
}

type BugSubmission struct {
	ID             domain.ID
	Access         Access
	Input          domain.BugInput
	IdempotencyKey string
	RequestSHA256  [32]byte
	ObservedAt     time.Time
}

type ID = domain.ID

type RuntimeStore interface {
	ContactSubmit(context.Context, ContactSubmission) (domain.ContactMessage, bool, error)
	ContactsList(context.Context, Access, domain.ContactFilter) (domain.ContactPage, error)
	ContactStatsRead(context.Context, Access, time.Time) (domain.ContactStats, error)
	ContactGet(context.Context, Access, domain.ID) (domain.ContactMessage, error)
	ContactUpdate(context.Context, Access, domain.ContactUpdate) (domain.ContactMessage, bool, error)
	BugCreate(context.Context, BugSubmission) (domain.BugReport, bool, error)
	BugsList(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	BugStatsRead(context.Context, Access, domain.ID) (domain.BugStats, error)
	BugGet(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
	BugUpdate(context.Context, Access, domain.BugUpdate) (domain.BugReport, bool, error)
}

type ContactPersistence interface {
	SubmitContact(context.Context, ContactSubmission) (domain.ContactMessage, bool, error)
	ListContacts(context.Context, Access, domain.ContactFilter) (domain.ContactPage, error)
	ReadContactStats(context.Context, Access, time.Time) (domain.ContactStats, error)
	GetContact(context.Context, Access, domain.ID) (domain.ContactMessage, error)
	UpdateContact(context.Context, Access, domain.ContactUpdate) (domain.ContactMessage, bool, error)
}

type ReporterBugPersistence interface {
	CreateBug(context.Context, BugSubmission) (domain.BugReport, bool, error)
	ListMyBugs(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	GetMyBug(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
}

type TenantBugPersistence interface {
	ListTenantBugs(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	ReadTenantBugStats(context.Context, Access, domain.ID) (domain.BugStats, error)
	GetTenantBug(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
	UpdateTenantBug(context.Context, Access, domain.BugUpdate) (domain.BugReport, bool, error)
}

type PlatformBugPersistence interface {
	ListPlatformBugs(context.Context, Access, domain.BugFilter) (domain.BugPage, error)
	ReadPlatformBugStats(context.Context, Access, domain.ID) (domain.BugStats, error)
	GetPlatformBug(context.Context, Access, domain.ID, domain.ID) (domain.BugReport, error)
	UpdatePlatformBug(context.Context, Access, domain.BugUpdate) (domain.BugReport, bool, error)
}

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewV7() (domain.ID, error) }

func Available(value any) bool {
	if value == nil {
		return false
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !reflected.IsNil()
	default:
		return true
	}
}

func ValidCommandEvidence(key string, digest [32]byte) bool {
	return len(key) >= 16 && len(key) <= 200 && digest != [32]byte{}
}

func NormalizeError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{ErrAuthentication, ErrForbidden, ErrInvalid,
		ErrNotFound, ErrConflict, ErrRateLimited, ErrIdempotencyMismatch, ErrRevisionStale} {
		if errors.Is(err, known) {
			return known
		}
	}
	return ErrUnavailable
}
