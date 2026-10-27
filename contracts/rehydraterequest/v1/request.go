package rehydraterequestv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	Schema               = "rehydrate-request.v1"
	MaximumDocumentBytes = 64 * 1024
	maximumAge           = 24 * time.Hour
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Request struct {
	Schema                string `json:"schema"`
	RecoverySetID         string `json:"recovery_set_id"`
	ExpectedEnvironmentID string `json:"expected_environment_id"`
	TargetDisposition     string `json:"target_disposition"`
	WaiverReason          string `json:"waiver_reason"`
	ProbeSHA256           string `json:"probe_sha256"`
	RecoveryKitSHA256     string `json:"recovery_kit_sha256"`
	DestructiveConfirmed  bool   `json:"destructive_confirmed"`
	RequestedAt           string `json:"requested_at"`
}

func Build(request Request, kit []byte, now time.Time) ([]byte, error) {
	kitDigest := sha256.Sum256(kit)
	request.Schema = Schema
	request.RecoveryKitSHA256 = hex.EncodeToString(kitDigest[:])
	raw, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("rehydrate request encoding failed")
	}
	document, err := canonicaljsonv1.CanonicalizeObject(
		raw, MaximumDocumentBytes,
	)
	if err != nil || !request.valid(kit, now) {
		return nil, fmt.Errorf("rehydrate request is invalid")
	}
	return document, nil
}

func Parse(document, kit []byte, now time.Time) (Request, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumDocumentBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) || len(kit) == 0 {
		return Request{}, fmt.Errorf("rehydrate request is invalid")
	}
	var request Request
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF ||
		!request.valid(kit, now) {
		return Request{}, fmt.Errorf("rehydrate request is invalid")
	}
	return request, nil
}

func (request Request) valid(kit []byte, now time.Time) bool {
	requestedAt, err := time.Parse(time.RFC3339Nano, request.RequestedAt)
	kitDigest := sha256.Sum256(kit)
	common := request.Schema == Schema &&
		validUUID(request.RecoverySetID) &&
		validUUID(request.ExpectedEnvironmentID) &&
		digestPattern.MatchString(request.ProbeSHA256) &&
		request.RecoveryKitSHA256 == hex.EncodeToString(kitDigest[:]) &&
		err == nil && canonicalTime(requestedAt) &&
		canonicalTime(now) && !requestedAt.After(now) &&
		now.Sub(requestedAt) <= maximumAge
	if !common {
		return false
	}
	switch request.TargetDisposition {
	case "empty":
		return request.WaiverReason == "empty_no_database_or_policy" &&
			!request.DestructiveConfirmed
	case "unreadable":
		return request.WaiverReason == "unreadable_integrity_probe" &&
			request.DestructiveConfirmed
	default:
		return false
	}
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
