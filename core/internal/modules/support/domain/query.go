package domain

import "time"

type Page struct {
	Number int
	Size   int
}

func (page Page) Valid() bool {
	return page.Number >= 1 && page.Size >= 1 && page.Size <= 100 &&
		(page.Number-1)*page.Size <= 5000
}

type ContactFilter struct {
	Status      ContactStatus
	ContactType string
	Page        Page
}

type BugFilter struct {
	TenantID   ID
	ReporterID ID
	Status     BugStatus
	Severity   BugSeverity
	Category   BugCategory
	Page       Page
}

type ContactStats struct {
	Total, New, Read, Replied, Archived, ThisMonth int64
}

type BugStats struct {
	Total, Open, InProgress, Resolved, Closed, CriticalOpen int64
	ByCategory                                              map[string]int64
	BySeverity                                              map[string]int64
}

type ContactPage struct {
	Items []ContactMessage
	Total int64
	Page  Page
}

type BugPage struct {
	Items []BugReport
	Total int64
	Page  Page
}

type ContactUpdate struct {
	ID               ID
	Status           *ContactStatus
	Notes            *string
	ExpectedRevision int64
	ActorID          ID
	IdempotencyKey   string
	RequestSHA256    [32]byte
	ObservedAt       time.Time
}

type BugUpdate struct {
	ID, TenantID, ActorID  ID
	Status                 *BugStatus
	Priority               *BugPriority
	Resolution, AdminNotes *string
	ExpectedRevision       int64
	IdempotencyKey         string
	RequestSHA256          [32]byte
	ObservedAt             time.Time
}
