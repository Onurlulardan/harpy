package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("support value is invalid")

var emailPattern = regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)+$`)

type ContactStatus string

const (
	ContactNew      ContactStatus = "new"
	ContactRead     ContactStatus = "read"
	ContactReplied  ContactStatus = "replied"
	ContactArchived ContactStatus = "archived"
)

type ContactInput struct {
	Name, Email, Company, Phone, EmployeeCount string
	Message, ContactType, Source, Honeypot     string
}

type ContactMessage struct {
	ID                                                  ID
	Name, Email, Company, Phone, EmployeeCount, Message string
	ContactType, Source, Notes                          string
	Status                                              ContactStatus
	RepliedAt                                           *time.Time
	RepliedBy                                           *ID
	Revision                                            int64
	CreatedAt, UpdatedAt                                time.Time
}

func NormalizeContact(input ContactInput) (ContactInput, error) {
	input.Name = clean(input.Name)
	input.Email = strings.ToLower(clean(input.Email))
	input.Company = clean(input.Company)
	input.Phone = clean(input.Phone)
	input.EmployeeCount = clean(input.EmployeeCount)
	input.Message = clean(input.Message)
	input.ContactType = clean(input.ContactType)
	input.Source = clean(input.Source)
	if !bounded(input.Name, 1, 200) || !bounded(input.Email, 3, 255) ||
		!emailPattern.MatchString(input.Email) || !bounded(input.Company, 0, 200) ||
		!bounded(input.Phone, 0, 50) || !bounded(input.Message, 1, 2000) ||
		!bounded(input.Source, 0, 50) || !validEmployeeCount(input.EmployeeCount) ||
		!validContactType(input.ContactType) {
		return ContactInput{}, ErrInvalid
	}
	return input, nil
}

func ValidContactStatus(value ContactStatus) bool {
	switch value {
	case ContactNew, ContactRead, ContactReplied, ContactArchived:
		return true
	default:
		return false
	}
}

func validContactType(value string) bool {
	switch value {
	case "general", "demo_request", "pricing_inquiry", "technical_question",
		"partnership", "enterprise_plan", "on_premise":
		return true
	default:
		return false
	}
}

func validEmployeeCount(value string) bool {
	switch value {
	case "", "1-10", "11-50", "51-200", "201-500", "500+":
		return true
	default:
		return false
	}
}

func clean(value string) string { return strings.TrimSpace(value) }

func bounded(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) &&
		utf8.RuneCountInString(value) >= minimum && len([]byte(value)) <= maximum
}
