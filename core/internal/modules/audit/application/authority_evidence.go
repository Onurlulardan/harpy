package application

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var authorityNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,127}$`)

type Evidence struct {
	actorID, tenantID                                      uuid.UUID
	tokenVersion, authorizationVersion, authorizationClock int64
	permission, operation                                  string
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

func (evidence Evidence) Authorizes(scope platformtx.TenantScope, permission string) bool {
	return evidence.valid() && evidence.actorID == scope.ActorID() &&
		evidence.tenantID == scope.TenantID() && evidence.operation == scope.OperationID() &&
		evidence.permission == permission
}

func (evidence Evidence) Digest() [sha256.Size]byte {
	if !evidence.valid() {
		return [sha256.Size]byte{}
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte("harpy/audit/authority-evidence/v1\x00"))
	_, _ = digest.Write(evidence.actorID[:])
	_, _ = digest.Write(evidence.tenantID[:])
	var versions [24]byte
	binary.BigEndian.PutUint64(versions[0:8], uint64(evidence.tokenVersion))
	binary.BigEndian.PutUint64(versions[8:16], uint64(evidence.authorizationVersion))
	binary.BigEndian.PutUint64(versions[16:24], uint64(evidence.authorizationClock))
	_, _ = digest.Write(versions[:])
	_, _ = digest.Write([]byte(evidence.permission + "\x00" + evidence.operation))
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func (evidence Evidence) Versions() (int64, int64, int64, bool) {
	if !evidence.valid() {
		return 0, 0, 0, false
	}
	return evidence.tokenVersion, evidence.authorizationVersion,
		evidence.authorizationClock, true
}

func (evidence Evidence) valid() bool {
	return evidence.issued && evidence.actorID != uuid.Nil && evidence.tenantID != uuid.Nil &&
		evidence.tokenVersion > 0 && evidence.authorizationVersion > 0 &&
		evidence.authorizationClock > 0 && authorityNamePattern.MatchString(evidence.permission) &&
		authorityNamePattern.MatchString(evidence.operation)
}
