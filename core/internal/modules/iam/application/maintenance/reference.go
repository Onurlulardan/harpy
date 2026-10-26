package maintenance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func RecurringReference(tenantID uuid.UUID, kind Kind) (string, string, error) {
	if !stableID(tenantID) || !kind.Recurring() {
		return "", "", ErrClaimInvalid
	}
	reference := fmt.Sprintf(
		"urn:harpy:iam:maintenance:%s:%s:v1", tenantID, kind,
	)
	return reference, digestReference(reference), nil
}

func ProfileReference(
	tenantID uuid.UUID, planID uuid.UUID, page int64,
) (string, string, error) {
	if !stableID(tenantID) || !stableID(planID) || page < 1 {
		return "", "", ErrClaimInvalid
	}
	reference := fmt.Sprintf(
		"urn:harpy:iam:profile-backfill:%s:%s:%d:v1",
		tenantID, planID, page,
	)
	return reference, digestReference(reference), nil
}

func parseProfileReference(value string) (uuid.UUID, uuid.UUID, int64, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 8 || parts[0] != "urn" || parts[1] != "harpy" ||
		parts[2] != "iam" || parts[3] != "profile-backfill" || parts[7] != "v1" {
		return uuid.Nil, uuid.Nil, 0, false
	}
	tenantID, tenantErr := uuid.Parse(parts[4])
	planID, planErr := uuid.Parse(parts[5])
	page, pageErr := strconv.ParseInt(parts[6], 10, 64)
	return tenantID, planID, page,
		tenantErr == nil && planErr == nil && stableID(tenantID) &&
			stableID(planID) && pageErr == nil && page > 0
}

func digestReference(reference string) string {
	digest := sha256.Sum256([]byte(reference))
	return hex.EncodeToString(digest[:])
}

func stableID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}
