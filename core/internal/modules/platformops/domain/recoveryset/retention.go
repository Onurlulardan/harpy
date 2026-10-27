package recoveryset

import (
	"sort"
	"strconv"
	"time"
)

type RetentionCandidate struct {
	ID          string
	ProtectedAt time.Time
	LegalHold   bool
}

type RetentionDecision struct {
	Keep   []string
	Delete []string
}

func ApplyRetention(values []RetentionCandidate) RetentionDecision {
	ordered := append([]RetentionCandidate(nil), values...)
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].ProtectedAt.Equal(ordered[right].ProtectedAt) {
			return ordered[left].ID < ordered[right].ID
		}
		return ordered[left].ProtectedAt.After(ordered[right].ProtectedAt)
	})
	keep := make(map[string]bool, 10)
	eligible := make([]RetentionCandidate, 0, len(ordered))
	for _, candidate := range ordered {
		if candidate.LegalHold {
			keep[candidate.ID] = true
		} else {
			eligible = append(eligible, candidate)
		}
	}
	for index := 0; index < len(eligible) && index < 2; index++ {
		keep[eligible[index].ID] = true
	}
	selectBuckets(eligible, keep, 5, weekBucket)
	selectBuckets(eligible, keep, 3, monthBucket)
	decision := RetentionDecision{}
	for _, candidate := range ordered {
		if keep[candidate.ID] {
			decision.Keep = append(decision.Keep, candidate.ID)
		} else {
			decision.Delete = append(decision.Delete, candidate.ID)
		}
	}
	return decision
}

func selectBuckets(
	values []RetentionCandidate,
	keep map[string]bool,
	maximum int,
	bucket func(time.Time) string,
) {
	selected := make(map[string]bool, maximum)
	for _, candidate := range values {
		if keep[candidate.ID] {
			continue
		}
		key := bucket(candidate.ProtectedAt)
		if selected[key] || len(selected) == maximum {
			continue
		}
		selected[key], keep[candidate.ID] = true, true
	}
}

func weekBucket(value time.Time) string {
	year, week := value.ISOWeek()
	return strconv.Itoa(year) + "-" + strconv.Itoa(week)
}

func monthBucket(value time.Time) string { return value.UTC().Format("2006-01") }
