package adoption

import (
	"bytes"
	"encoding/json"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type bugSourceDocument struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	ReporterID       uuid.UUID       `json:"reporter_id"`
	ReporterName     string          `json:"reporter_name"`
	ReporterEmail    string          `json:"reporter_email"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Category         string          `json:"category"`
	Severity         string          `json:"severity"`
	Priority         string          `json:"priority"`
	PageURL          string          `json:"page_url"`
	UserAgent        string          `json:"user_agent"`
	BrowserInfo      json.RawMessage `json:"browser_info"`
	ConsoleErrors    json.RawMessage `json:"console_errors"`
	NetworkInfo      json.RawMessage `json:"network_info"`
	Screenshots      json.RawMessage `json:"screenshots"`
	StepsToReproduce string          `json:"steps_to_reproduce"`
	Status           string          `json:"status"`
	Resolution       string          `json:"resolution"`
	AdminNotes       string          `json:"admin_notes"`
	ResolvedAt       *time.Time      `json:"resolved_at"`
	ClosedAt         *time.Time      `json:"closed_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type bugTargetDocument struct {
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
	Screenshots      json.RawMessage    `json:"screenshots"`
	StepsToReproduce string             `json:"steps_to_reproduce"`
	Status           domain.BugStatus   `json:"status"`
	Resolution       string             `json:"resolution"`
	AdminNotes       string             `json:"admin_notes"`
	ResolvedAt       *time.Time         `json:"resolved_at"`
	ClosedAt         *time.Time         `json:"closed_at"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

func bugRecords(run Run, page BugPage) ([]Record, error) {
	if !validPage(page.ManifestSHA256, run.ManifestSHA256, page.After, run.CursorID,
		len(page.Rows), page.Done) {
		return nil, ErrSourceDrift
	}
	records := make([]Record, len(page.Rows))
	previous := run.CursorID
	for index, row := range page.Rows {
		if !orderedSourceID(previous, row.ID) || row.TenantID != run.TenantID {
			return nil, ErrSourceDrift
		}
		record, err := bugRecord(row)
		if err != nil {
			return nil, err
		}
		records[index], previous = record, row.ID
	}
	return records, nil
}

func BugSourcePageSHA256(rows []BugSourceRow) ([32]byte, error) {
	records := make([]Record, len(rows))
	previous := uuid.Nil
	for index, row := range rows {
		if !orderedSourceID(previous, row.ID) {
			return [32]byte{}, ErrSourceDrift
		}
		record, err := bugRecord(row)
		if err != nil {
			return [32]byte{}, err
		}
		records[index], previous = record, row.ID
	}
	return PageSHA256(records), nil
}

func bugRecord(row BugSourceRow) (Record, error) {
	createdAt, created := storedTime(row.CreatedAt)
	updatedAt, updated := storedTime(row.UpdatedAt)
	resolvedAt, resolved := storedOptionalTime(row.ResolvedAt)
	closedAt, closed := storedOptionalTime(row.ClosedAt)
	if row.ID == uuid.Nil || row.TenantID == uuid.Nil || row.ReporterID == uuid.Nil ||
		!created || !updated || !resolved || !closed || updatedAt.Before(createdAt) ||
		row.DeletedAt != nil || !validBugState(row) ||
		!boundedText(row.ReporterName, 0, 200) || !boundedText(row.ReporterEmail, 0, 255) ||
		!boundedText(row.Resolution, 0, 8000) || !boundedText(row.AdminNotes, 0, 8000) {
		return Record{}, ErrSourceDrift
	}
	input, err := domain.NormalizeBug(domain.BugInput{Title: row.Title,
		Description: row.Description, Category: domain.BugCategory(row.Category),
		Severity: domain.BugSeverity(row.Severity), PageURL: row.PageURL,
		UserAgent: row.UserAgent, BrowserInfo: row.BrowserInfo,
		ConsoleErrors: row.ConsoleErrors, NetworkInfo: row.NetworkInfo,
		StepsToReproduce: row.StepsToReproduce}, canonicaljsonv1.Canonicalize)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	screenshots, err := canonicalArray(row.Screenshots, 64<<10, 50)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	source := bugSourceDocument{ID: row.ID, TenantID: row.TenantID,
		ReporterID: row.ReporterID, ReporterName: row.ReporterName,
		ReporterEmail: row.ReporterEmail, Title: row.Title, Description: row.Description,
		Category: row.Category, Severity: row.Severity, Priority: row.Priority,
		PageURL: row.PageURL, UserAgent: row.UserAgent, BrowserInfo: row.BrowserInfo,
		ConsoleErrors: row.ConsoleErrors, NetworkInfo: row.NetworkInfo,
		Screenshots: row.Screenshots, StepsToReproduce: row.StepsToReproduce,
		Status: row.Status, Resolution: row.Resolution, AdminNotes: row.AdminNotes,
		ResolvedAt: resolvedAt, ClosedAt: closedAt, CreatedAt: createdAt, UpdatedAt: updatedAt}
	_, sourceSHA, err := canonicalDocument(source, 192<<10)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	target := bugTargetDocument{ReporterID: row.ReporterID,
		ReporterName: row.ReporterName, ReporterEmail: row.ReporterEmail,
		Title: input.Title, Description: input.Description, Category: input.Category,
		Severity: input.Severity, Priority: domain.BugPriority(row.Priority),
		PageURL: input.PageURL, UserAgent: input.UserAgent, BrowserInfo: input.BrowserInfo,
		ConsoleErrors: input.ConsoleErrors, NetworkInfo: input.NetworkInfo,
		Screenshots: screenshots, StepsToReproduce: input.StepsToReproduce,
		Status: domain.BugStatus(row.Status), Resolution: row.Resolution,
		AdminNotes: row.AdminNotes, ResolvedAt: resolvedAt, ClosedAt: closedAt,
		CreatedAt: createdAt, UpdatedAt: updatedAt}
	document, _, err := canonicalDocument(target, 192<<10)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	return Record{SourceID: row.ID, SourceSHA256: hexSHA(sourceSHA), Target: document}, nil
}

func canonicalArray(value json.RawMessage, maximum, maxItems int) (json.RawMessage, error) {
	if len(bytes.TrimSpace(value)) == 0 {
		value = json.RawMessage(`[]`)
	}
	document, err := canonicaljsonv1.Canonicalize(value, maximum)
	if err != nil || len(document) == 0 || document[0] != '[' {
		return nil, ErrInvalid
	}
	var items []json.RawMessage
	if json.Unmarshal(document, &items) != nil || len(items) > maxItems {
		return nil, ErrInvalid
	}
	return document, nil
}
