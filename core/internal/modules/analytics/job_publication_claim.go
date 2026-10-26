package analytics

import (
	"fmt"
	"strconv"
	"strings"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type analyticsPublicationJobClaim struct {
	releaseID  uuid.UUID
	generation uint64
	state      string
}

func parseAnalyticsPublicationJobClaim(
	claim platformasync.DatabaseJobClaim, kind string,
) (analyticsPublicationJobClaim, error) {
	if err := validateAnalyticsJobClaim(claim, kind); err != nil {
		return analyticsPublicationJobClaim{}, err
	}
	parts := strings.Split(claim.PayloadReference(), "/")
	if len(parts) != 4 || parts[0] != "analytics-publication-repair" ||
		(parts[3] != "active" && parts[3] != "revoked") {
		return analyticsPublicationJobClaim{}, fmt.Errorf(
			"Analytics publication job reference is invalid")
	}
	releaseID, idErr := uuid.Parse(parts[1])
	generation, generationErr := strconv.ParseUint(parts[2], 10, 63)
	if idErr != nil || !validJobUUID(releaseID) || generationErr != nil || generation < 1 {
		return analyticsPublicationJobClaim{}, fmt.Errorf(
			"Analytics publication job identity is invalid")
	}
	return analyticsPublicationJobClaim{releaseID: releaseID,
		generation: generation, state: parts[3]}, nil
}
