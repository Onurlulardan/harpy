package postgres

import (
	"encoding/json"
	"time"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type bugDocument struct {
	ID               uuid.UUID          `json:"id"`
	TenantID         uuid.UUID          `json:"tenant_id"`
	TenantSlug       string             `json:"tenant_slug"`
	TenantName       string             `json:"tenant_name"`
	ReporterID       uuid.UUID          `json:"reporter_id"`
	ReporterName     string             `json:"reporter_name"`
	ReporterEmail    string             `json:"reporter_email"`
	Title            string             `json:"title"`
	Description      string             `json:"description"`
	Category         domain.BugCategory `json:"category"`
	Severity         domain.BugSeverity `json:"severity"`
	Priority         domain.BugPriority `json:"priority"`
	PageURL          string             `json:"page_url"`
	UserAgent        string             `json:"user_agent"`
	BrowserInfo      json.RawMessage    `json:"browser_info"`
	ConsoleErrors    json.RawMessage    `json:"console_errors"`
	NetworkInfo      json.RawMessage    `json:"network_info"`
	StepsToReproduce string             `json:"steps_to_reproduce"`
	Status           domain.BugStatus   `json:"status"`
	Resolution       string             `json:"resolution"`
	AdminNotes       string             `json:"admin_notes"`
	ResolvedAt       *time.Time         `json:"resolved_at"`
	ClosedAt         *time.Time         `json:"closed_at"`
	Revision         int64              `json:"revision"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func decodeBug(document []byte) (domain.BugReport, error) {
	var wire bugDocument
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil ||
		wire.ID == uuid.Nil || wire.ID.Version() != 7 || wire.TenantID == uuid.Nil ||
		wire.ReporterID == uuid.Nil || wire.Revision < 1 ||
		!domain.ValidBugStatus(wire.Status) || !domain.ValidBugPriority(wire.Priority) ||
		!validStoredTime(wire.CreatedAt) || !validStoredTime(wire.UpdatedAt) ||
		wire.UpdatedAt.Before(wire.CreatedAt) || !storedText(wire.TenantSlug, 1, 100) ||
		!storedText(wire.TenantName, 1, 200) || !storedText(wire.ReporterName, 0, 200) ||
		!storedText(wire.ReporterEmail, 0, 255) || !storedText(wire.Resolution, 0, 8000) ||
		!storedText(wire.AdminNotes, 0, 8000) {
		return domain.BugReport{}, errDocumentInvalid
	}
	input, err := domain.NormalizeBug(domain.BugInput{Title: wire.Title,
		Description: wire.Description, Category: wire.Category, Severity: wire.Severity,
		StepsToReproduce: wire.StepsToReproduce, PageURL: wire.PageURL,
		UserAgent: wire.UserAgent, BrowserInfo: wire.BrowserInfo,
		ConsoleErrors: wire.ConsoleErrors, NetworkInfo: wire.NetworkInfo}, canonicaljsonv1.Canonicalize)
	if err != nil {
		return domain.BugReport{}, errDocumentInvalid
	}
	return domain.BugReport{ID: domain.ID(wire.ID), TenantID: domain.ID(wire.TenantID),
		ReporterID: domain.ID(wire.ReporterID), TenantSlug: wire.TenantSlug,
		TenantName: wire.TenantName, ReporterName: wire.ReporterName,
		ReporterEmail: wire.ReporterEmail, Title: input.Title,
		Description: input.Description, Category: input.Category, Severity: input.Severity,
		Priority: wire.Priority, PageURL: input.PageURL,
		StepsToReproduce: input.StepsToReproduce, Status: wire.Status,
		Resolution: wire.Resolution, AdminNotes: wire.AdminNotes, Revision: wire.Revision,
		ResolvedAt: wire.ResolvedAt, ClosedAt: wire.ClosedAt,
		CreatedAt: wire.CreatedAt, UpdatedAt: wire.UpdatedAt}, nil
}

func decodeBugPage(document []byte) ([]domain.BugReport, error) {
	var values []json.RawMessage
	if json.Unmarshal(document, &values) != nil || len(values) > 100 {
		return nil, errDocumentInvalid
	}
	result := make([]domain.BugReport, len(values))
	for index, value := range values {
		report, err := decodeBug(value)
		if err != nil {
			return nil, err
		}
		result[index] = report
	}
	return result, nil
}

func storedText(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && len(value) >= minimum && len([]byte(value)) <= maximum
}
