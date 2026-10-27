package campaign

import (
	"sort"
	"time"
)

type Selection struct {
	version   Version
	discounts []Discount
}

func (selection Selection) Version() Version {
	return selection.version
}

func (selection Selection) Discounts() []Discount {
	return append([]Discount(nil), selection.discounts...)
}

func SelectActive(instant time.Time, candidates []Version) (*Selection, error) {
	at, err := canonicalInstant(instant)
	if err != nil {
		return nil, err
	}
	eligible := make([]Version, 0, len(candidates))
	for _, candidate := range candidates {
		if !validIdentifier(candidate.id) || !validIdentifier(candidate.campaignID) ||
			candidate.start.IsZero() {
			return nil, ErrVersion
		}
		if candidate.eligible(at) {
			eligible = append(eligible, candidate)
		}
	}
	if len(eligible) == 0 {
		return nil, nil
	}
	sort.Slice(eligible, func(left, right int) bool {
		if eligible[left].priority != eligible[right].priority {
			return eligible[left].priority > eligible[right].priority
		}
		return compareIdentifier(eligible[left].id, eligible[right].id) < 0
	})
	selected := eligible[0]
	discounts := make([]Discount, 0, len(selected.discounts))
	for _, discount := range selected.discounts {
		if discount.eligible(at) {
			discounts = append(discounts, discount)
		}
	}
	sort.Slice(discounts, func(left, right int) bool {
		if discounts[left].priority != discounts[right].priority {
			return discounts[left].priority < discounts[right].priority
		}
		return compareIdentifier(discounts[left].id, discounts[right].id) < 0
	})
	return &Selection{version: selected, discounts: discounts}, nil
}
