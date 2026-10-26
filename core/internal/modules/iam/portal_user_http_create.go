package iam

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *portalUserHTTPApplication) CreatePortalUser(
	ctx context.Context, claims accesstoken.Claims, input PortalUserCreateInput,
) (PortalUserView, error) {
	if !identityIdempotencyPattern.MatchString(input.IdempotencyKey) ||
		input.SubjectID == uuid.Nil || !portalSubjectKeyPattern.MatchString(input.FormSlug) ||
		!portalSubjectKeyPattern.MatchString(input.SubjectType) {
		return PortalUserView{}, ErrPortalUserHTTPInvalid
	}
	ids := newPortalUserIDs(claims, input.IdempotencyKey)
	var view portaluser.View
	err := application.inCommand(ctx, claims, "iam.portal_user.create", "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			resolution, resolveErr := application.resolver.ResolvePortalSubject(scoped, scope,
				PortalSubjectResolutionInput{ActorID: claims.Subject(), TenantID: claims.Tenant(),
					FormSlug: input.FormSlug, SubjectType: input.SubjectType,
					SubjectRecordID: input.SubjectID})
			if resolveErr != nil || !resolution.valid(PortalSubjectResolutionInput{
				FormSlug: input.FormSlug, SubjectType: input.SubjectType,
			}) {
				return ErrPortalUserHTTPInvalid
			}
			command, prepareErr := application.preparePortalCreate(
				scoped, scope, claims, input, resolution, ids, permissionID,
			)
			if prepareErr != nil {
				return prepareErr
			}
			result, createErr := application.users.Create(scoped, scope, command)
			if createErr != nil {
				return createErr
			}
			if finishErr := application.finishPortalUserChange(
				scoped, scope, command.Evidence, result.ID,
				"iam.portal_user.create", result.Replayed,
			); finishErr != nil {
				return finishErr
			}
			view, createErr = application.users.Read(scoped, scope, result.ID)
			return createErr
		})
	if err != nil {
		return PortalUserView{}, err
	}
	return portalUserView(view), nil
}

func (application *portalUserHTTPApplication) preparePortalCreate(
	ctx context.Context, scope platformtx.TenantScope, claims identityEvidenceSource,
	input PortalUserCreateInput, resolution PortalSubjectResolution,
	ids portalUserIDs, permissionID uuid.UUID,
) (portaluser.CreateCommand, error) {
	portalValue, err := identifier.NormalizePortalV1(resolution.Identifier)
	if err != nil {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	emailValue, err := identifier.NormalizeEmailV1(resolution.Email)
	if err != nil {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	if input.Email != "" {
		requested, normalizeErr := identifier.NormalizeEmailV1(input.Email)
		if normalizeErr != nil || requested.Normalized() != emailValue.Normalized() {
			return portaluser.CreateCommand{}, ErrPortalUserHTTPConflict
		}
	}
	firstName, lastName := strings.TrimSpace(input.FirstName), strings.TrimSpace(input.LastName)
	if firstName == "" {
		firstName = strings.TrimSpace(resolution.FirstName)
	}
	if lastName == "" {
		lastName = strings.TrimSpace(resolution.LastName)
	}
	if !validIdentityText(firstName, 200) || !validIdentityText(lastName, 200) {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	roles, valid := canonicalPortalRoles(resolution.RoleNames)
	if !valid {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	status := input.Status
	if status == "" {
		status = resolution.DefaultStatus
	}
	if status == "" {
		status = "PENDING"
	}
	if status != "PENDING" && status != "INACTIVE" {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	profile, err := application.portalProfile(ctx, scope, firstName, lastName)
	if err != nil {
		return portaluser.CreateCommand{}, err
	}
	requestSHA, err := identityRequestSHA(struct {
		Email, Portal, FirstName, LastName, FormSlug, SubjectType, Status string
		SubjectID, FormID, VersionID, PublicationID, FieldID              uuid.UUID
		Roles                                                             []string
		Contract                                                          [32]byte
	}{emailValue.Normalized(), portalValue.Normalized(), firstName, lastName,
		input.FormSlug, input.SubjectType, status, input.SubjectID, resolution.FormID,
		resolution.FormVersionID, resolution.PublicationID, resolution.IdentifierFieldID,
		roles, resolution.ContractSHA256})
	if err != nil {
		return portaluser.CreateCommand{}, ErrPortalUserHTTPInvalid
	}
	return portaluser.CreateCommand{UserID: ids.user, Profile: profile, SetupState: status,
		Portal: application.identifier(portalValue, ids.portalIdentifier),
		Email:  application.identifier(emailValue, ids.emailIdentifier),
		Subject: portaluser.Subject{BindingID: ids.binding, AccessModeID: ids.portalAccess,
			FormID: resolution.FormID, FormVersionID: resolution.FormVersionID,
			RecordID: input.SubjectID, IdentifierFieldID: resolution.IdentifierFieldID,
			Type: input.SubjectType, ContractSHA256: resolution.ContractSHA256},
		RoleNames: roles, RoleAssignmentIDs: portalRoleAssignmentIDs(ids.operation, roles),
		Evidence: portalMutationEvidence(claims, permissionID, ids.operation, requestSHA)}, nil
}
