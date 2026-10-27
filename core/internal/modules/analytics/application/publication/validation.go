package publication

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"regexp"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	"github.com/google/uuid"
)

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
func requestDigest(parts ...[]byte) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(part)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func transitionDigest(operation string, identity ScopeIdentity, request TransitionRequest) [32]byte {
	revision := make([]byte, 8)
	binary.BigEndian.PutUint64(revision, request.ExpectedRevision)
	return requestDigest([]byte(operation), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.ReleaseID[:], revision)
}

func activationDigest(identity ScopeIdentity, request ActivateRequest) [32]byte {
	revision := make([]byte, 8)
	ttl := make([]byte, 8)
	binary.BigEndian.PutUint64(revision, request.ExpectedRevision)
	binary.BigEndian.PutUint64(ttl, uint64(request.CapabilityTTL))
	return requestDigest([]byte(OperationActivate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.ReleaseID[:], revision, ttl)
}

func validateSnapshot(value Snapshot) error {
	id, err := analyticspublication.NewID([16]byte(value.ReleaseID))
	if err != nil || !value.ResourceKind.Valid() || !validUUIDv7(value.ResourceVersionID) ||
		!value.Audience.Valid() || !value.State.Valid() || !value.CachePolicy.Valid() ||
		value.ManifestSHA256 == [32]byte{} || value.RequesterID == uuid.Nil ||
		value.Revision == 0 || value.Generation == 0 || !validTime(value.CreatedAt) ||
		!validTime(value.UpdatedAt) || value.UpdatedAt.Before(value.CreatedAt) {
		return ErrUnavailable
	}
	if value.Audience != analyticspublication.AudienceTenant && value.PublicID == [16]byte{} &&
		value.State == analyticspublication.StateActive {
		return ErrUnavailable
	}
	if value.Audience == analyticspublication.AudienceCapability && value.State == analyticspublication.StateActive {
		if value.CapabilityExpiresAt == nil || !validTime(*value.CapabilityExpiresAt) {
			return ErrUnavailable
		}
	}
	_, err = analyticspublication.NewRelease(id, [16]byte(value.ResourceVersionID),
		value.Audience, value.CachePolicy, value.ManifestSHA256, [16]byte(value.RequesterID))
	return mapDomainError(err)
}

func mapDomainError(err error) error {
	if err == nil {
		return nil
	}
	if err == analyticspublication.ErrConflict {
		return ErrConflict
	}
	return ErrUnavailable
}

func validPublicBody(value []byte) bool {
	return len(value) > 0 && len(value) <= 1<<20 && json.Valid(value)
}
