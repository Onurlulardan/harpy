package billingv1adoptionapp

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

func campaignPriorities(sources []SourceCampaign) (map[uuid.UUID]int32, map[uuid.UUID]bool) {
	type value struct {
		id        uuid.UUID
		createdAt time.Time
	}
	ordered := make([]value, 0, len(sources))
	counts := map[string]int{}
	for _, source := range sources {
		if source.CreatedAt == nil {
			continue
		}
		ordered = append(ordered, value{id: source.ID, createdAt: *source.CreatedAt})
		counts[source.CreatedAt.Format(time.RFC3339Nano)]++
	}
	sort.Slice(ordered, func(i, j int) bool {
		if !ordered[i].createdAt.Equal(ordered[j].createdAt) {
			return ordered[i].createdAt.Before(ordered[j].createdAt)
		}
		return ordered[i].id.String() < ordered[j].id.String()
	})
	priorities := make(map[uuid.UUID]int32, len(ordered))
	ambiguous := make(map[uuid.UUID]bool)
	for position, item := range ordered {
		priorities[item.id] = int32(position + 1)
		ambiguous[item.id] = counts[item.createdAt.Format(time.RFC3339Nano)] > 1
	}
	return priorities, ambiguous
}
