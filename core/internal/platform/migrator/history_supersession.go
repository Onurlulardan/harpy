package migrator

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

type historySupersession struct {
	version       int64
	owner         string
	scope         string
	origin        postgresv1.DatabaseOrigin
	currentPath   string
	currentSHA256 string
	superseded    []historyIdentity
}

type historyIdentity struct {
	path   string
	sha256 string
}

func indexHistorySupersessions(
	selected []postgresv1.SelectedMigration,
	values []migration.HistorySupersession,
) (map[int64]historySupersession, error) {
	selectedByVersion := make(map[int64]postgresv1.SelectedMigration, len(selected))
	for _, current := range selected {
		selectedByVersion[current.Version] = current
	}
	result := make(map[int64]historySupersession, len(values))
	for _, value := range values {
		current, exists := selectedByVersion[value.Version()]
		if !exists || value.Owner() != current.Owner || value.Scope() != current.Scope ||
			value.Origin() != current.Origin || value.CurrentPath() != current.Path ||
			value.CurrentSHA256() != current.SHA256 || value.SupersededPath() == current.Path ||
			value.SupersededSHA256() == current.SHA256 {
			return nil, fmt.Errorf("migration history supersession differs from selection")
		}
		indexed, exists := result[value.Version()]
		if !exists {
			indexed = historySupersession{
				version: value.Version(), owner: value.Owner(), scope: value.Scope(),
				origin: value.Origin(), currentPath: value.CurrentPath(),
				currentSHA256: value.CurrentSHA256(),
			}
		} else if !indexed.matchesCurrent(current) {
			return nil, fmt.Errorf("migration history supersession version differs")
		}
		for _, identity := range indexed.superseded {
			if identity.path == value.SupersededPath() {
				return nil, fmt.Errorf("migration history supersession alias is duplicate")
			}
		}
		indexed.superseded = append(indexed.superseded, historyIdentity{
			path: value.SupersededPath(), sha256: value.SupersededSHA256(),
		})
		result[value.Version()] = indexed
	}
	return result, nil
}

func (value historySupersession) matchesCurrent(
	current postgresv1.SelectedMigration,
) bool {
	return value.version == current.Version && value.owner == current.Owner &&
		value.scope == current.Scope && value.origin == current.Origin &&
		value.currentPath == current.Path && value.currentSHA256 == current.SHA256
}
