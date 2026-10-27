package licenseauthority

import (
	"time"

	"github.com/google/uuid"
)

const maximumSafeInteger = int64(9_007_199_254_740_991)

type parsedPayload struct {
	wire                                     wirePayload
	licenseID, organizationID, environmentID uuid.UUID
	issuedAt, notBefore, refreshAfter        time.Time
	grantValidUntil, graceUntil              time.Time
	commercialExpiresAt                      time.Time
}

func parsePayload(document []byte) (parsedPayload, error) {
	var parsed parsedPayload
	if strictDecode(document, &parsed.wire) != nil || !validPayloadConstants(parsed.wire) {
		return parsedPayload{}, ErrInvalidDocument
	}
	var valid bool
	if parsed.licenseID, valid = parseExactUUID(parsed.wire.LicenseID); !valid {
		return parsedPayload{}, ErrInvalidDocument
	}
	if parsed.organizationID, valid = parseExactUUID(
		parsed.wire.LicenseeOrganizationID,
	); !valid {
		return parsedPayload{}, ErrInvalidDocument
	}
	if parsed.environmentID, valid = parseExactUUID(parsed.wire.EnvironmentID); !valid {
		return parsedPayload{}, ErrInvalidDocument
	}
	instants := []*time.Time{
		&parsed.issuedAt, &parsed.notBefore, &parsed.refreshAfter,
		&parsed.grantValidUntil, &parsed.graceUntil, &parsed.commercialExpiresAt,
	}
	values := []string{
		parsed.wire.IssuedAt, parsed.wire.NotBefore, parsed.wire.RefreshAfter,
		parsed.wire.GrantValidUntil, parsed.wire.GraceUntil,
		parsed.wire.CommercialExpiresAt,
	}
	for index := range values {
		if *instants[index], valid = parseExactInstant(values[index]); !valid {
			return parsedPayload{}, ErrInvalidDocument
		}
	}
	if !validTimeline(parsed) || !validLimits(parsed.wire.Limits) {
		return parsedPayload{}, ErrInvalidDocument
	}
	return parsed, nil
}

func validPayloadConstants(value wirePayload) bool {
	return value.Schema == SnapshotSchema && value.Audience == SnapshotAudience &&
		value.Product == SnapshotProduct && value.DeploymentMode == "onpremise" &&
		validText(value.Issuer, 256) && validDelivery(value.DeliveryKind) &&
		validStatus(value.Status) && value.Sequence > 0 &&
		channelPattern.MatchString(value.ReleaseChannel)
}

func validTimeline(value parsedPayload) bool {
	return !value.issuedAt.After(value.notBefore) &&
		!value.notBefore.After(value.refreshAfter) &&
		!value.refreshAfter.After(value.grantValidUntil) &&
		!value.grantValidUntil.After(value.graceUntil) &&
		!value.notBefore.After(value.commercialExpiresAt)
}

func validLimits(value wireLimits) bool {
	limits := []int64{
		value.MaxUsers, value.MaxForms, value.MaxWorkflows,
		value.MaxAPIEndpoints, value.MaxDashboards, value.MaxTenants,
	}
	for _, limit := range limits {
		if limit < 0 || limit > maximumSafeInteger {
			return false
		}
	}
	return true
}
