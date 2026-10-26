package iam

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	"github.com/google/uuid"
)

const rbacStepUpNamespace = "harpy/iam/rbac-step-up/v1"

func rbacOperationID(
	claims accesstoken.Claims,
	command string,
	idempotencyKey string,
) uuid.UUID {
	return identityOperationID(claims, command, idempotencyKey)
}

func rbacStepUpPurpose(command string, target uuid.UUID) [32]byte {
	targetWire := ""
	if target != uuid.Nil && !strings.HasSuffix(command, ".create") {
		targetWire = target.String()
	}
	return sha256.Sum256([]byte(
		rbacStepUpNamespace + "|" + command + "|" + targetWire,
	))
}

func (application *rbacHTTPApplication) stepUp(
	claims accesstoken.Claims,
	command string,
	target uuid.UUID,
) (rbacdomain.StepUpEvidence, bool) {
	purpose, present := claims.StepUpPurposeHash()
	now := application.now().UTC().Truncate(time.Second)
	authTime := time.Unix(claims.AuthenticationTimeUnix(), 0).UTC()
	if !present || purpose != rbacStepUpPurpose(command, target) ||
		claims.Assurance() != accesstoken.AssuranceStepUp ||
		authTime.After(now) || now.Sub(authTime) > 10*time.Minute {
		return rbacdomain.StepUpEvidence{}, false
	}
	evidence := rbacdomain.StepUpEvidence{
		PurposeSHA256: purpose, AuthTime: authTime,
	}
	for _, method := range claims.Methods() {
		evidence.AMRPassword = evidence.AMRPassword ||
			method == accesstoken.AuthMethodPassword
		evidence.AMRWebAuthn = evidence.AMRWebAuthn ||
			method == accesstoken.AuthMethodWebAuthn
		evidence.AMRTOTP = evidence.AMRTOTP ||
			method == accesstoken.AuthMethodTOTP
	}
	return evidence, evidence.Validate() == nil
}

func rbacMutationEvidence(
	claims accesstoken.Claims,
	permissionID uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
) rbacdomain.MutationEvidence {
	return rbacdomain.MutationEvidence{
		OperationID: rbacdomain.ID(operationID), RequestSHA256: requestSHA,
		ActorID:                      rbacdomain.ID(claims.Subject()),
		RequiredPermissionID:         rbacdomain.ID(permissionID),
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
}

func rbacQueryEvidence(
	evidence identityEvidenceSource,
	permissionID uuid.UUID,
) rbacdomain.QueryEvidence {
	return rbacdomain.QueryEvidence{
		ActorID:              rbacdomain.ID(evidence.Subject()),
		RequiredPermissionID: rbacdomain.ID(permissionID),
	}
}

func rbacAssignmentItems(
	operationID uuid.UUID,
	actionIDs []uuid.UUID,
) []rbacdomain.AssignmentItem {
	ids := identitySortedIDs(actionIDs)
	items := make([]rbacdomain.AssignmentItem, len(ids))
	for index, actionID := range ids {
		items[index] = rbacdomain.AssignmentItem{
			Ordinal: index + 1, ActionID: rbacdomain.ID(actionID),
			PermissionID: rbacdomain.ID(identityChildID(
				operationID, fmt.Sprintf("permission/%s", actionID),
			)),
			BindingID: rbacdomain.ID(identityChildID(
				operationID, fmt.Sprintf("binding/%s", actionID),
			)),
		}
	}
	return items
}
