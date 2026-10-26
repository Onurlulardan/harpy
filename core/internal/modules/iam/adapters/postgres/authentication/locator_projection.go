package postgres

import "time"

type locatorAdvance struct {
	selector     []byte
	expected     *int64
	revision     *int64
	sourceSHA256 []byte
	state        *string
	projectedAt  *time.Time
	retainUntil  *time.Time
}

func (projection locatorAdvance) valid() bool {
	return len(projection.selector) == 16 && projection.expected != nil &&
		projection.revision != nil && *projection.expected > 0 &&
		*projection.revision == *projection.expected+1 &&
		len(projection.sourceSHA256) == 32 && projection.state != nil &&
		(*projection.state == "consumed" || *projection.state == "revoked" ||
			*projection.state == "expired") && projection.projectedAt != nil &&
		!projection.projectedAt.IsZero() && projection.retainUntil != nil &&
		!projection.retainUntil.IsZero()
}

type locatorCreate struct {
	selector      []byte
	revision      *int64
	sourceSHA256  []byte
	state         *string
	issuedAt      *time.Time
	projectedAt   *time.Time
	familyExpires *time.Time
	retainUntil   *time.Time
}

func (projection locatorCreate) valid() bool {
	return len(projection.selector) == 16 && projection.revision != nil &&
		*projection.revision == 1 && len(projection.sourceSHA256) == 32 &&
		projection.state != nil && *projection.state == "active" &&
		projection.issuedAt != nil && !projection.issuedAt.IsZero() &&
		projection.projectedAt != nil && !projection.projectedAt.IsZero() &&
		projection.familyExpires != nil && !projection.familyExpires.IsZero() &&
		projection.retainUntil != nil && !projection.retainUntil.IsZero()
}

type rotationRow struct {
	outcome   string
	issuedAt  *time.Time
	advance   locatorAdvance
	successor locatorCreate
}
