package authority

import (
	"crypto/sha256"
	"regexp"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,127}$`)

type Evidence struct {
	actorID, tenantID                                      uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	permission, operation                                  string
	authenticatedAt                                        time.Time
	authorizationPermissionID                              uuid.UUID
	authorizationEvidenceSHA256, stepUpPurposeSHA256       [sha256.Size]byte
	worker, issued                                         bool
}

func NewEvidence(actorID, tenantID uuid.UUID, tokenVersion, authorizationVersion,
	authorizationClock int64, permission, operation string,
	authenticatedAt time.Time,
) (Evidence, error) {
	evidence := Evidence{actorID: actorID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission,
		operation: operation, authenticatedAt: authenticatedAt, issued: true}
	if !evidence.valid() {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
}

func NewEvidenceWithAuthorizationReceipt(
	actorID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission, operation string,
	authenticatedAt time.Time,
	authorizationPermissionID uuid.UUID,
	authorizationEvidenceSHA256, stepUpPurposeSHA256 [sha256.Size]byte,
) (Evidence, error) {
	evidence, err := NewEvidence(
		actorID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, permission, operation, authenticatedAt,
	)
	if err != nil || authorizationPermissionID == uuid.Nil ||
		authorizationEvidenceSHA256 == [sha256.Size]byte{} ||
		(!authenticatedAt.IsZero() && stepUpPurposeSHA256 == [sha256.Size]byte{}) ||
		(authenticatedAt.IsZero() && stepUpPurposeSHA256 != [sha256.Size]byte{}) {
		return Evidence{}, ErrInvalid
	}
	evidence.authorizationPermissionID = authorizationPermissionID
	evidence.authorizationEvidenceSHA256 = authorizationEvidenceSHA256
	evidence.stepUpPurposeSHA256 = stepUpPurposeSHA256
	return evidence, nil
}

func NewWorkerEvidence(actorID, tenantID uuid.UUID, permission, operation string) (Evidence, error) {
	evidence := Evidence{actorID: actorID, tenantID: tenantID,
		permission: permission, operation: operation,
		worker: true, issued: true}
	if !evidence.valid() {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
}

func (evidence Evidence) Authorizes(scope platformtx.TenantScope, permission string) bool {
	return evidence.valid() && evidence.actorID == scope.ActorID() &&
		evidence.tenantID == scope.TenantID() && evidence.operation == scope.OperationID() &&
		evidence.permission == permission
}

func (evidence Evidence) FreshStepUp(now time.Time, maximumAge time.Duration) (time.Time, bool) {
	if !evidence.valid() || evidence.worker || evidence.authenticatedAt.IsZero() ||
		now.IsZero() || maximumAge <= 0 || evidence.authenticatedAt.After(now) ||
		now.Sub(evidence.authenticatedAt) > maximumAge {
		return time.Time{}, false
	}
	return evidence.authenticatedAt, true
}

func (evidence Evidence) AuthorizationReceipt() (
	int64, int64, int64, uuid.UUID, [sha256.Size]byte, [sha256.Size]byte, bool,
) {
	valid := evidence.valid() && !evidence.worker &&
		evidence.authorizationPermissionID != uuid.Nil &&
		evidence.authorizationEvidenceSHA256 != [sha256.Size]byte{}
	return evidence.tokenVersion, evidence.authorizationVersion,
		evidence.authorizationClock, evidence.authorizationPermissionID,
		evidence.authorizationEvidenceSHA256, evidence.stepUpPurposeSHA256, valid
}

func (evidence Evidence) valid() bool {
	if !evidence.issued || evidence.actorID == uuid.Nil || evidence.tenantID == uuid.Nil ||
		!namePattern.MatchString(evidence.permission) || !namePattern.MatchString(evidence.operation) {
		return false
	}
	if evidence.worker {
		return evidence.tokenVersion == 0 && evidence.authorizationVersion == 0 &&
			evidence.authorizationClock == 0 && evidence.authenticatedAt.IsZero() &&
			evidence.permission == evidence.operation
	}
	return evidence.tokenVersion > 0 && evidence.authorizationVersion > 0 &&
		evidence.authorizationClock > 0 &&
		(evidence.authenticatedAt.IsZero() || evidence.authenticatedAt.Location() == time.UTC)
}
