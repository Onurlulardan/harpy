package v1adoption

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type legacyIndex struct {
	boards      map[domain.ID]LegacyBoard
	columns     map[domain.ID]LegacyColumn
	memberships map[domain.ID]LegacyMembership
	tasks       map[domain.ID]LegacyTask
	comments    map[domain.ID]LegacyComment
	calendar    map[domain.ID]LegacyCalendar
	sources     map[SourceKind]map[domain.ID]SourceRecord
}

func indexLegacy(records []SourceRecord) (legacyIndex, error) {
	index := legacyIndex{boards: map[domain.ID]LegacyBoard{}, columns: map[domain.ID]LegacyColumn{},
		memberships: map[domain.ID]LegacyMembership{}, tasks: map[domain.ID]LegacyTask{},
		comments: map[domain.ID]LegacyComment{}, calendar: map[domain.ID]LegacyCalendar{},
		sources: map[SourceKind]map[domain.ID]SourceRecord{}}
	for _, record := range records {
		if index.sources[record.Kind] == nil {
			index.sources[record.Kind] = map[domain.ID]SourceRecord{}
		}
		index.sources[record.Kind][record.ID] = record
		if err := index.decode(record); err != nil {
			return legacyIndex{}, err
		}
	}
	return index, nil
}

func (index *legacyIndex) decode(record SourceRecord) error {
	switch record.Kind {
	case KindBoard:
		value := LegacyBoard{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.boards[value.ID] = value
	case KindColumn:
		value := LegacyColumn{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.columns[value.ID] = value
	case KindMembership:
		value := LegacyMembership{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.memberships[value.ID] = value
	case KindTask:
		value := LegacyTask{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.tasks[value.ID] = value
	case KindComment:
		value := LegacyComment{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.comments[value.ID] = value
	case KindCalendar:
		value := LegacyCalendar{}
		if json.Unmarshal(record.Document, &value) != nil || value.ID != record.ID {
			return ports.ErrUnavailable
		}
		index.calendar[value.ID] = value
	default:
		return ports.ErrInvalid
	}
	return nil
}

func (index legacyIndex) source(kind SourceKind, id domain.ID) SourceRecord {
	return index.sources[kind][id]
}
