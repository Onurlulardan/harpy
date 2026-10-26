package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"sort"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type portalUserIDs struct {
	operation, user, portalIdentifier, emailIdentifier uuid.UUID
	portalAccess, binding, consoleAccess, consoleRole  uuid.UUID
}

func newPortalUserIDs(claims identityEvidenceSource, key string) portalUserIDs {
	operation := identityOperationID(claims, "iam.portal_user.create", key)
	return portalUserIDs{operation: operation,
		user:             identityChildID(operation, "user"),
		portalIdentifier: identityChildID(operation, "portal-identifier"),
		emailIdentifier:  identityChildID(operation, "email-identifier"),
		portalAccess:     identityChildID(operation, "portal-access-mode"),
		binding:          identityChildID(operation, "portal-binding"),
		consoleAccess:    identityChildID(operation, "console-access-mode"),
		consoleRole:      identityChildID(operation, "console-role-assignment")}
}

func portalMutationEvidence(
	claims identityEvidenceSource, permissionID, operationID uuid.UUID, requestSHA [32]byte,
) portaluser.Evidence {
	return portaluser.Evidence{OperationID: operationID, RequestSHA256: requestSHA,
		ActorID: claims.Subject(), PermissionID: permissionID,
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion()}
}

func (application *portalUserHTTPApplication) identifier(
	value identifier.Value, id uuid.UUID,
) portaluser.Identifier {
	digest := value.NormalizedSHA256()
	mac := hmac.New(sha256.New, application.directory.material[:])
	_, _ = mac.Write([]byte(value.Normalized()))
	var lookup [32]byte
	copy(lookup[:], mac.Sum(nil))
	return portaluser.Identifier{ID: id, Display: value.Display(),
		Normalized: value.Normalized(), NormalizedSHA256: digest,
		LookupKeyID: application.directory.id, LookupHMAC: lookup}
}

func canonicalPortalRoles(values []string) ([]string, bool) {
	roles := append([]string(nil), values...)
	for index := range roles {
		roles[index] = strings.TrimSpace(roles[index])
		if roles[index] == "" || len(roles[index]) > 255 {
			return nil, false
		}
	}
	sort.Strings(roles)
	for index := 1; index < len(roles); index++ {
		if roles[index] == roles[index-1] {
			return nil, false
		}
	}
	return roles, len(roles) > 0 && len(roles) <= 32
}

func portalRoleAssignmentIDs(operation uuid.UUID, roles []string) []uuid.UUID {
	result := make([]uuid.UUID, len(roles))
	for index, role := range roles {
		result[index] = identityChildID(operation, "portal-role:"+role)
	}
	return result
}

func (application *portalUserHTTPApplication) finishPortalUserChange(
	ctx context.Context, scope platformtx.TenantScope, evidence portaluser.Evidence,
	userID uuid.UUID, command string, replayed bool,
) error {
	if replayed {
		return nil
	}
	change := changeapp.NewChange(identityfacade.MutationEvidence{
		OperationID:                  identityfacade.IdentityID(evidence.OperationID),
		RequestSHA256:                evidence.RequestSHA256,
		ActorID:                      identityfacade.IdentityID(evidence.ActorID),
		PermissionID:                 identityfacade.IdentityID(evidence.PermissionID),
		ExpectedTokenVersion:         evidence.ExpectedTokenVersion,
		ExpectedAuthorizationVersion: evidence.ExpectedAuthorizationVersion,
	}, identityfacade.IdentityID(userID), "user", command)
	if change.Validate() != nil || application.changes.WriteChange(ctx, scope, change) != nil {
		return portaluser.ErrUnavailable
	}
	return nil
}

func portalUserView(value portaluser.View) PortalUserView {
	return PortalUserView{ID: value.ID, Email: value.Email, Status: value.Status,
		AccountType: value.AccountType, BindingID: value.BindingID,
		SubjectType: value.SubjectType, SubjectID: value.SubjectRecordID,
		Revision: value.Revision, SetupRequired: value.SetupRequired,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func portalProfileValid(document identityfacade.ProfileDocument) bool {
	return document.Validate(identityfacade.ProfileUser) == nil
}
