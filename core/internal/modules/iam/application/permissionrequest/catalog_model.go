package permissionrequest

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var catalogKeyPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`,
)

type CatalogKind string

const (
	CatalogKindAll    CatalogKind = "all"
	CatalogKindForm   CatalogKind = "form"
	CatalogKindSystem CatalogKind = "system"
)

type CatalogQuery struct {
	Kind   CatalogKind
	Search string
	Limit  int
	After  uuid.UUID
}

type MissingCatalogQuery struct {
	CatalogQuery
	ActionKey string
}

type CatalogResource struct {
	ID          uuid.UUID
	Key         string
	DisplayName string
	Description string
	Kind        CatalogKind
}

type CatalogAction struct {
	ID          uuid.UUID
	Key         string
	DisplayName string
	Description string
}

type CatalogResourcePage struct {
	Items []CatalogResource
	Next  uuid.UUID
}

type CatalogActionPage struct {
	Items []CatalogAction
	Next  uuid.UUID
}

func (query CatalogQuery) Validate() error {
	if query.Kind != CatalogKindAll && query.Kind != CatalogKindForm &&
		query.Kind != CatalogKindSystem {
		return ErrInvalid
	}
	if query.Limit < 1 || query.Limit > 100 || len(query.Search) > 128 ||
		query.Search != strings.TrimSpace(query.Search) {
		return ErrInvalid
	}
	return nil
}

func (query MissingCatalogQuery) Validate() error {
	if query.CatalogQuery.Validate() != nil || !validCatalogKey(query.ActionKey) {
		return ErrInvalid
	}
	return nil
}

func validCatalogKey(value string) bool {
	return len(value) >= 1 && len(value) <= 100 && catalogKeyPattern.MatchString(value)
}

func (resource CatalogResource) Validate() error {
	if resource.ID == uuid.Nil || !validCatalogKey(resource.Key) ||
		resource.DisplayName == "" || resource.DisplayName != strings.TrimSpace(resource.DisplayName) ||
		(resource.Kind != CatalogKindForm && resource.Kind != CatalogKindSystem) {
		return ErrInvalid
	}
	return nil
}

func (action CatalogAction) Validate() error {
	if action.ID == uuid.Nil || !validCatalogKey(action.Key) ||
		action.DisplayName == "" || action.DisplayName != strings.TrimSpace(action.DisplayName) {
		return ErrInvalid
	}
	return nil
}
