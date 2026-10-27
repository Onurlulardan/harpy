package postgres

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

type contactStatsDocument struct {
	Total     int64 `json:"total"`
	New       int64 `json:"new"`
	Read      int64 `json:"read"`
	Replied   int64 `json:"replied"`
	Archived  int64 `json:"archived"`
	ThisMonth int64 `json:"this_month"`
}

func decodeContactStats(document []byte) (domain.ContactStats, error) {
	var wire contactStatsDocument
	if json.Unmarshal(document, &wire) != nil || wire.Total < 0 || wire.New < 0 ||
		wire.Read < 0 || wire.Replied < 0 || wire.Archived < 0 || wire.ThisMonth < 0 ||
		wire.New+wire.Read+wire.Replied+wire.Archived != wire.Total {
		return domain.ContactStats{}, errDocumentInvalid
	}
	return domain.ContactStats{Total: wire.Total, New: wire.New, Read: wire.Read,
		Replied: wire.Replied, Archived: wire.Archived, ThisMonth: wire.ThisMonth}, nil
}

type bugStatsDocument struct {
	Total        int64            `json:"total"`
	Open         int64            `json:"open"`
	InProgress   int64            `json:"in_progress"`
	Resolved     int64            `json:"resolved"`
	Closed       int64            `json:"closed"`
	CriticalOpen int64            `json:"critical_open"`
	ByCategory   map[string]int64 `json:"by_category"`
	BySeverity   map[string]int64 `json:"by_severity"`
}

func decodeBugStats(document []byte) (domain.BugStats, error) {
	var wire bugStatsDocument
	if json.Unmarshal(document, &wire) != nil || wire.Total < 0 || wire.Open < 0 ||
		wire.InProgress < 0 || wire.Resolved < 0 || wire.Closed < 0 ||
		wire.CriticalOpen < 0 || wire.CriticalOpen > wire.Open ||
		!validCounts(wire.ByCategory, wire.Total) || !validCounts(wire.BySeverity, wire.Total) {
		return domain.BugStats{}, errDocumentInvalid
	}
	return domain.BugStats{Total: wire.Total, Open: wire.Open,
		InProgress: wire.InProgress, Resolved: wire.Resolved, Closed: wire.Closed,
		CriticalOpen: wire.CriticalOpen, ByCategory: wire.ByCategory,
		BySeverity: wire.BySeverity}, nil
}

func validCounts(values map[string]int64, total int64) bool {
	var sum int64
	for key, value := range values {
		if key == "" || value < 0 {
			return false
		}
		sum += value
	}
	return sum == total
}
