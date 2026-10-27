package authority

import (
	"crypto/sha256"
	"encoding/binary"
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
	stepUp                                                 bool
	issued                                                 bool
}

func NewEvidence(
	actorID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission, operation string,
) (Evidence, error) {
	evidence := Evidence{actorID: actorID, tenantID: tenantID,
		tokenVersion: tokenVersion, authorizationVersion: authorizationVersion,
		authorizationClock: authorizationClock, permission: permission,
		operation: operation, issued: true}
	if !evidence.valid() {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
}

func NewStepUpEvidence(
	actorID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission, operation string,
	authenticatedAt time.Time,
) (Evidence, error) {
	evidence, err := NewEvidence(actorID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, permission, operation)
	if err != nil || authenticatedAt.IsZero() || authenticatedAt.Nanosecond() != 0 {
		return Evidence{}, ErrInvalid
	}
	evidence.authenticatedAt = authenticatedAt.UTC()
	evidence.stepUp = true
	return evidence, nil
}

func (evidence Evidence) Authorizes(
	scope platformtx.TenantScope, permission string,
) bool {
	return evidence.valid() && evidence.actorID == scope.ActorID() &&
		evidence.tenantID == scope.TenantID() && evidence.operation == scope.OperationID() &&
		evidence.permission == permission
}

func (evidence Evidence) Versions() (int64, int64, int64, bool) {
	if !evidence.valid() {
		return 0, 0, 0, false
	}
	return evidence.tokenVersion, evidence.authorizationVersion,
		evidence.authorizationClock, true
}

func (evidence Evidence) Digest() [sha256.Size]byte {
	if !evidence.valid() {
		return [sha256.Size]byte{}
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/analytics/authority-evidence/v1\x00"))
	_, _ = hash.Write(evidence.actorID[:])
	_, _ = hash.Write(evidence.tenantID[:])
	var versions [24]byte
	binary.BigEndian.PutUint64(versions[0:8], uint64(evidence.tokenVersion))
	binary.BigEndian.PutUint64(versions[8:16], uint64(evidence.authorizationVersion))
	binary.BigEndian.PutUint64(versions[16:24], uint64(evidence.authorizationClock))
	_, _ = hash.Write(versions[:])
	var authentication [8]byte
	binary.BigEndian.PutUint64(authentication[:], uint64(evidence.authenticatedAt.Unix()))
	_, _ = hash.Write(authentication[:])
	if evidence.stepUp {
		_, _ = hash.Write([]byte{1})
	} else {
		_, _ = hash.Write([]byte{0})
	}
	_, _ = hash.Write([]byte(evidence.permission + "\x00" + evidence.operation))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (evidence Evidence) FreshStepUp(now time.Time, maximumAge time.Duration) (time.Time, bool) {
	if !evidence.valid() || !evidence.stepUp || evidence.authenticatedAt.IsZero() ||
		now.IsZero() || now.Nanosecond() != 0 || maximumAge <= 0 ||
		evidence.authenticatedAt.After(now) || now.Sub(evidence.authenticatedAt) > maximumAge {
		return time.Time{}, false
	}
	return evidence.authenticatedAt, true
}

func (evidence Evidence) valid() bool {
	return evidence.issued && evidence.actorID != uuid.Nil && evidence.tenantID != uuid.Nil &&
		evidence.tokenVersion > 0 && evidence.authorizationVersion > 0 &&
		evidence.authorizationClock > 0 && namePattern.MatchString(evidence.permission) &&
		namePattern.MatchString(evidence.operation) &&
		(!evidence.stepUp || !evidence.authenticatedAt.IsZero())
}
