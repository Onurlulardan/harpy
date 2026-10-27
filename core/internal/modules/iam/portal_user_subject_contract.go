package iam

import (
	"context"
	"regexp"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var portalSubjectKeyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)

type PortalSubjectResolutionInput struct {
	ActorID, TenantID, SubjectRecordID uuid.UUID
	FormSlug, SubjectType              string
}

type PortalSubjectResolution struct {
	FormID, FormVersionID, PublicationID uuid.UUID
	IdentifierFieldID                    uuid.UUID
	Identifier, Email                    string
	FirstName, LastName                  string
	ProfileFormSlug, DefaultStatus       string
	RoleNames                            []string
	ContractSHA256                       [32]byte
}

func (value PortalSubjectResolution) valid(input PortalSubjectResolutionInput) bool {
	if value.FormID == uuid.Nil || value.FormVersionID == uuid.Nil ||
		value.PublicationID == uuid.Nil || value.IdentifierFieldID == uuid.Nil ||
		value.Identifier == "" || value.Email == "" || value.FirstName == "" ||
		value.LastName == "" || value.ProfileFormSlug != input.FormSlug ||
		value.ContractSHA256 == [32]byte{} || len(value.RoleNames) == 0 ||
		len(value.RoleNames) > 32 {
		return false
	}
	return value.DefaultStatus == "" || value.DefaultStatus == "PENDING" ||
		value.DefaultStatus == "INACTIVE"
}

type PortalSubjectResolver interface {
	ResolvePortalSubject(
		context.Context, platformtx.TenantScope, PortalSubjectResolutionInput,
	) (PortalSubjectResolution, error)
}
