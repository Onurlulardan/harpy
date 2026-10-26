package campaign

import "time"

type Version struct {
	id         Identifier
	campaignID Identifier
	priority   int32
	active     bool
	published  bool
	start      time.Time
	end        *time.Time
	discounts  []Discount
}

func NewVersion(
	id Identifier,
	campaignID Identifier,
	priority int32,
	active bool,
	published bool,
	start time.Time,
	end *time.Time,
	discounts []Discount,
) (Version, error) {
	canonicalStart, err := canonicalInstant(start)
	if err != nil || !validIdentifier(id) || !validIdentifier(campaignID) {
		return Version{}, ErrVersion
	}
	canonicalFinish, err := canonicalEnd(end, canonicalStart)
	if err != nil {
		return Version{}, ErrVersion
	}
	copyDiscounts := append([]Discount(nil), discounts...)
	for _, discount := range copyDiscounts {
		if !validIdentifier(discount.id) || discount.start.IsZero() {
			return Version{}, ErrVersion
		}
	}
	return Version{id: id, campaignID: campaignID, priority: priority,
		active: active, published: published, start: canonicalStart,
		end: canonicalFinish, discounts: copyDiscounts}, nil
}

func (version Version) Identifier() Identifier { return version.id }
func (version Version) CampaignID() Identifier { return version.campaignID }
func (version Version) Priority() int32        { return version.priority }

func (version Version) eligible(instant time.Time) bool {
	return version.active && version.published &&
		contains(version.start, version.end, instant)
}
