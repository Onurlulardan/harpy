package adoption

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var TransformSHA256 = sha256.Sum256([]byte("harpy.usage.adoption.route-template.v1"))

type AdoptedRecord struct {
	SourceID       string `json:"source_id"`
	TargetID       string `json:"target_id"`
	EnvironmentID  string `json:"environment_id"`
	TenantID       string `json:"tenant_id"`
	ActorID        string `json:"actor_id"`
	SourceIdentity string `json:"source_identity"`
	SourceSHA256   string `json:"source_sha256"`
	FactDigest     string `json:"fact_digest"`
	Method         string `json:"method"`
	RoutePattern   string `json:"route_pattern"`
	StatusCode     int    `json:"status_code"`
	DurationMicros int64  `json:"duration_micros"`
	OccurredAt     string `json:"occurred_at"`
}

func TransformPage(manifest Manifest, page SourcePage) ([]AdoptedRecord, [32]byte, error) {
	if manifest.Validate() != nil || page.Validate(100) != nil || manifest.Kind != LegacyTrace {
		return nil, [32]byte{}, domain.ErrInvalid
	}
	result := make([]AdoptedRecord, 0, len(page.Records))
	targetDocument := make([]byte, 0, len(page.Records)*96)
	for _, source := range page.Records {
		route, err := CanonicalRoute(source.Path)
		if err != nil {
			return nil, [32]byte{}, err
		}
		targetID := stableV7(manifest, source.SourceID, source.OccurredAt)
		sourceIdentity := sourceIdentity(manifest, source.SourceID)
		observation, err := domain.NormalizeObservation(domain.Observation{
			ID: domain.ID(targetID), EnvironmentID: domain.ID(source.EnvironmentID),
			TenantID: domain.ID(source.TenantID), ActorID: domain.ID(source.ActorID),
			SourceIdentity: sourceIdentity,
			Method:         strings.ToUpper(strings.TrimSpace(source.Method)), RoutePattern: route,
			StatusCode: source.StatusCode, DurationMicros: source.DurationMicros,
			OccurredAt: source.OccurredAt.UTC(),
		})
		if err != nil {
			return nil, [32]byte{}, err
		}
		sourceSHA := sourceDigest(source)
		factDigest := observation.Digest()
		result = append(result, AdoptedRecord{SourceID: source.SourceID.String(),
			TargetID: targetID.String(), EnvironmentID: source.EnvironmentID.String(),
			TenantID: source.TenantID.String(), ActorID: source.ActorID.String(),
			SourceIdentity: hex.EncodeToString(sourceIdentity[:]),
			SourceSHA256:   hex.EncodeToString(sourceSHA[:]),
			FactDigest:     hex.EncodeToString(factDigest[:]), Method: observation.Method,
			RoutePattern: observation.RoutePattern, StatusCode: observation.StatusCode,
			DurationMicros: observation.DurationMicros,
			OccurredAt:     observation.OccurredAt.Format(time.RFC3339Nano)})
		targetDocument = append(targetDocument, source.SourceID[:]...)
		targetDocument = append(targetDocument, factDigest[:]...)
	}
	return result, sha256.Sum256(targetDocument), nil
}

func stableV7(manifest Manifest, sourceID uuid.UUID, occurredAt time.Time) uuid.UUID {
	seed := make([]byte, 0, 64)
	seed = append(seed, []byte(manifest.Schema+"."+manifest.Relation)...)
	seed = append(seed, sourceID[:]...)
	hash := sha256.Sum256(seed)
	millis := uint64(occurredAt.UTC().UnixMilli())
	var result uuid.UUID
	for index := 5; index >= 0; index-- {
		result[index] = byte(millis)
		millis >>= 8
	}
	copy(result[6:], hash[:10])
	result[6] = (result[6] & 0x0f) | 0x70
	result[8] = (result[8] & 0x3f) | 0x80
	return result
}

func sourceIdentity(manifest Manifest, sourceID uuid.UUID) [32]byte {
	var oid [4]byte
	binary.BigEndian.PutUint32(oid[:], manifest.RelationOID)
	return sha256.Sum256(append(append([]byte(manifest.Schema+"."+manifest.Relation), oid[:]...),
		sourceID[:]...))
}
