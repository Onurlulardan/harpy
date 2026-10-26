package postgres

import (
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

type contactDocument struct {
	ID            uuid.UUID            `json:"id"`
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
	Revision      int64                `json:"revision"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

func decodeContact(document []byte) (domain.ContactMessage, error) {
	var wire contactDocument
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil ||
		wire.ID == uuid.Nil || wire.ID.Version() != 7 || wire.Revision < 1 ||
		!domain.ValidContactStatus(wire.Status) || !validStoredTime(wire.CreatedAt) ||
		!validStoredTime(wire.UpdatedAt) || wire.UpdatedAt.Before(wire.CreatedAt) ||
		(wire.RepliedAt == nil) != (wire.RepliedBy == nil) {
		return domain.ContactMessage{}, errDocumentInvalid
	}
	input, err := domain.NormalizeContact(domain.ContactInput{Name: wire.Name,
		Email: wire.Email, Company: wire.Company, Phone: wire.Phone,
		EmployeeCount: wire.EmployeeCount, Message: wire.Message,
		ContactType: wire.ContactType, Source: wire.Source})
	if err != nil || len([]byte(wire.Notes)) > 8000 {
		return domain.ContactMessage{}, errDocumentInvalid
	}
	return domain.ContactMessage{ID: domain.ID(wire.ID), Name: input.Name, Email: input.Email,
		Company: input.Company, Phone: input.Phone, EmployeeCount: input.EmployeeCount,
		Message: input.Message, ContactType: input.ContactType, Source: input.Source,
		Status: wire.Status, Notes: wire.Notes, RepliedAt: wire.RepliedAt,
		RepliedBy: supportIDPointer(wire.RepliedBy), Revision: wire.Revision,
		CreatedAt: wire.CreatedAt, UpdatedAt: wire.UpdatedAt}, nil
}

func supportIDPointer(value *uuid.UUID) *domain.ID {
	if value == nil {
		return nil
	}
	id := domain.ID(*value)
	return &id
}

func decodeContactPage(document []byte) ([]domain.ContactMessage, error) {
	var values []json.RawMessage
	if json.Unmarshal(document, &values) != nil || len(values) > 100 {
		return nil, errDocumentInvalid
	}
	result := make([]domain.ContactMessage, len(values))
	for index, value := range values {
		message, err := decodeContact(value)
		if err != nil {
			return nil, err
		}
		result[index] = message
	}
	return result, nil
}
