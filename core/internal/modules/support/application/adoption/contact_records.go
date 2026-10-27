package adoption

import (
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type contactSourceDocument struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Company       string     `json:"company"`
	Phone         string     `json:"phone"`
	EmployeeCount string     `json:"employee_count"`
	Message       string     `json:"message"`
	ContactType   string     `json:"contact_type"`
	Status        string     `json:"status"`
	Source        string     `json:"source"`
	IPAddress     string     `json:"ip_address"`
	UserAgent     string     `json:"user_agent"`
	Notes         string     `json:"notes"`
	RepliedAt     *time.Time `json:"replied_at"`
	RepliedBy     *uuid.UUID `json:"replied_by"`
	CreatedAt     time.Time  `json:"created_at"`
}

type contactTargetDocument struct {
	Name          string               `json:"name"`
	Email         string               `json:"email"`
	Company       string               `json:"company"`
	Phone         string               `json:"phone"`
	EmployeeCount string               `json:"employee_count"`
	Message       string               `json:"message"`
	ContactType   string               `json:"contact_type"`
	Status        domain.ContactStatus `json:"status"`
	Source        string               `json:"source"`
	Notes         string               `json:"notes"`
	RepliedAt     *time.Time           `json:"replied_at"`
	RepliedBy     *uuid.UUID           `json:"replied_by"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

func contactRecords(run Run, page ContactPage) ([]Record, error) {
	if !validPage(page.ManifestSHA256, run.ManifestSHA256, page.After, run.CursorID,
		len(page.Rows), page.Done) {
		return nil, ErrSourceDrift
	}
	records := make([]Record, len(page.Rows))
	previous := run.CursorID
	for index, row := range page.Rows {
		if !orderedSourceID(previous, row.ID) {
			return nil, ErrSourceDrift
		}
		record, err := contactRecord(row)
		if err != nil {
			return nil, err
		}
		records[index], previous = record, row.ID
	}
	return records, nil
}

func ContactSourcePageSHA256(rows []ContactSourceRow) ([32]byte, error) {
	records := make([]Record, len(rows))
	previous := uuid.Nil
	for index, row := range rows {
		if !orderedSourceID(previous, row.ID) {
			return [32]byte{}, ErrSourceDrift
		}
		record, err := contactRecord(row)
		if err != nil {
			return [32]byte{}, err
		}
		records[index], previous = record, row.ID
	}
	return PageSHA256(records), nil
}

func contactRecord(row ContactSourceRow) (Record, error) {
	createdAt, valid := storedTime(row.CreatedAt)
	repliedAt, repliedValid := storedOptionalTime(row.RepliedAt)
	if row.ID == uuid.Nil || !valid || !repliedValid ||
		(row.RepliedAt == nil) != (row.RepliedBy == nil) ||
		!validContactStatus(row.Status) || !boundedText(row.Notes, 0, 8000) ||
		!boundedText(row.IPAddress, 0, 45) || !boundedText(row.UserAgent, 0, 8192) {
		return Record{}, ErrSourceDrift
	}
	input, err := domain.NormalizeContact(domain.ContactInput{Name: row.Name, Email: row.Email,
		Company: row.Company, Phone: row.Phone, EmployeeCount: row.EmployeeCount,
		Message: row.Message, ContactType: row.ContactType, Source: row.Source})
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	source := contactSourceDocument{ID: row.ID, Name: row.Name, Email: row.Email,
		Company: row.Company, Phone: row.Phone, EmployeeCount: row.EmployeeCount,
		Message: row.Message, ContactType: row.ContactType, Status: row.Status,
		Source: row.Source, IPAddress: row.IPAddress, UserAgent: row.UserAgent,
		Notes: row.Notes, RepliedAt: repliedAt, RepliedBy: row.RepliedBy, CreatedAt: createdAt}
	_, sourceSHA, err := canonicalDocument(source, 64<<10)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	target := contactTargetDocument{Name: input.Name, Email: input.Email,
		Company: input.Company, Phone: input.Phone, EmployeeCount: input.EmployeeCount,
		Message: input.Message, ContactType: input.ContactType,
		Status: domain.ContactStatus(row.Status), Source: input.Source, Notes: row.Notes,
		RepliedAt: repliedAt, RepliedBy: row.RepliedBy, CreatedAt: createdAt, UpdatedAt: createdAt}
	document, _, err := canonicalDocument(target, 64<<10)
	if err != nil {
		return Record{}, ErrSourceDrift
	}
	return Record{SourceID: row.ID, SourceSHA256: hexSHA(sourceSHA),
		Target: json.RawMessage(document)}, nil
}
