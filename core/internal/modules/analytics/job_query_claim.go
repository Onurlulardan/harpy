package analytics

import (
	"encoding/hex"
	"fmt"
	"strings"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type analyticsQueryJobClaim struct {
	queryID, versionID uuid.UUID
	manifest           [32]byte
}

func parseAnalyticsQueryJobClaim(
	claim platformasync.DatabaseJobClaim, kind, prefix string,
) (analyticsQueryJobClaim, error) {
	if err := validateAnalyticsJobClaim(claim, kind); err != nil {
		return analyticsQueryJobClaim{}, err
	}
	parts := strings.Split(claim.PayloadReference(), "/")
	if len(parts) != 4 || parts[0] != prefix {
		return analyticsQueryJobClaim{}, fmt.Errorf("Analytics query job reference is invalid")
	}
	queryID, queryErr := uuid.Parse(parts[1])
	versionID, versionErr := uuid.Parse(parts[2])
	digest, digestErr := hex.DecodeString(parts[3])
	if queryErr != nil || versionErr != nil || !validJobUUID(queryID) ||
		!validJobUUID(versionID) || digestErr != nil || len(digest) != 32 {
		return analyticsQueryJobClaim{}, fmt.Errorf("Analytics query job identity is invalid")
	}
	var manifest [32]byte
	copy(manifest[:], digest)
	if manifest == [32]byte{} {
		return analyticsQueryJobClaim{}, fmt.Errorf("Analytics query job manifest is invalid")
	}
	return analyticsQueryJobClaim{queryID: queryID,
		versionID: versionID, manifest: manifest}, nil
}
