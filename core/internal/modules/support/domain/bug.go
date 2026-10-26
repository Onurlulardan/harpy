package domain

import (
	"bytes"
	"encoding/json"
	"net/url"
	"time"
)

type JSONCanonicalizer func([]byte, int) ([]byte, error)

type BugStatus string
type BugCategory string
type BugSeverity string
type BugPriority string

const (
	BugOpen       BugStatus = "open"
	BugInProgress BugStatus = "in_progress"
	BugResolved   BugStatus = "resolved"
	BugClosed     BugStatus = "closed"
	BugWontFix    BugStatus = "wont_fix"
	BugDuplicate  BugStatus = "duplicate"
)

type BugInput struct {
	Title, Description, StepsToReproduce, PageURL, UserAgent string
	Category                                                 BugCategory
	Severity                                                 BugSeverity
	BrowserInfo, ConsoleErrors, NetworkInfo                  json.RawMessage
}

type BugReport struct {
	ID, TenantID, ReporterID                          ID
	TenantSlug, TenantName, ReporterName              string
	ReporterEmail, Title, Description                 string
	Category                                          BugCategory
	Severity                                          BugSeverity
	Priority                                          BugPriority
	PageURL, StepsToReproduce, Resolution, AdminNotes string
	Status                                            BugStatus
	Revision                                          int64
	ResolvedAt, ClosedAt                              *time.Time
	CreatedAt, UpdatedAt                              time.Time
}

func NormalizeBug(input BugInput, canonicalize JSONCanonicalizer) (BugInput, error) {
	if canonicalize == nil {
		return BugInput{}, ErrInvalid
	}
	input.Title, input.Description = clean(input.Title), clean(input.Description)
	input.StepsToReproduce, input.PageURL = clean(input.StepsToReproduce), clean(input.PageURL)
	input.UserAgent = clean(input.UserAgent)
	if input.Category == "" {
		input.Category = "bug"
	}
	if input.Severity == "" {
		input.Severity = "medium"
	}
	if !bounded(input.Title, 3, 500) || !bounded(input.Description, 10, 8000) ||
		!bounded(input.StepsToReproduce, 0, 8000) || !validPageURL(input.PageURL) ||
		!bounded(input.UserAgent, 0, 1024) || !ValidBugCategory(input.Category) ||
		!ValidBugSeverity(input.Severity) {
		return BugInput{}, ErrInvalid
	}
	var err error
	if input.BrowserInfo, err = canonicalContainer(input.BrowserInfo, '{', 16<<10, 0, canonicalize); err != nil {
		return BugInput{}, ErrInvalid
	}
	if input.ConsoleErrors, err = canonicalContainer(input.ConsoleErrors, '[', 32<<10, 50, canonicalize); err != nil {
		return BugInput{}, ErrInvalid
	}
	if input.NetworkInfo, err = canonicalContainer(input.NetworkInfo, '[', 32<<10, 50, canonicalize); err != nil {
		return BugInput{}, ErrInvalid
	}
	return input, nil
}

func ValidBugStatus(value BugStatus) bool {
	switch value {
	case BugOpen, BugInProgress, BugResolved, BugClosed, BugWontFix, BugDuplicate:
		return true
	default:
		return false
	}
}

func ValidBugCategory(value BugCategory) bool {
	switch value {
	case "bug", "feature_request", "performance", "ui_issue", "data_issue", "other":
		return true
	default:
		return false
	}
}

func ValidBugSeverity(value BugSeverity) bool {
	return value == "critical" || value == "high" || value == "medium" || value == "low"
}

func ValidBugPriority(value BugPriority) bool {
	return value == "urgent" || value == "high" || value == "normal" || value == "low"
}

func validPageURL(value string) bool {
	if !bounded(value, 0, 2048) || value == "" {
		return value == ""
	}
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func canonicalContainer(value json.RawMessage, prefix byte, limit, items int,
	canonicalize JSONCanonicalizer,
) (json.RawMessage, error) {
	if len(bytes.TrimSpace(value)) == 0 {
		if prefix == '{' {
			return json.RawMessage(`{}`), nil
		}
		return json.RawMessage(`[]`), nil
	}
	canonical, err := canonicalize(value, limit)
	if err != nil || len(canonical) == 0 || canonical[0] != prefix {
		return nil, ErrInvalid
	}
	if items > 0 {
		var elements []json.RawMessage
		if json.Unmarshal(canonical, &elements) != nil || len(elements) > items {
			return nil, ErrInvalid
		}
	}
	return canonical, nil
}
