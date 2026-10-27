package migration

import (
	"fmt"
	"sort"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type HistorySupersession struct {
	version          int64
	owner            string
	scope            string
	origin           postgresv1.DatabaseOrigin
	currentPath      string
	currentSHA256    string
	supersededPath   string
	supersededSHA256 string
}

func (value HistorySupersession) Version() int64                    { return value.version }
func (value HistorySupersession) Owner() string                     { return value.owner }
func (value HistorySupersession) Scope() string                     { return value.scope }
func (value HistorySupersession) Origin() postgresv1.DatabaseOrigin { return value.origin }
func (value HistorySupersession) CurrentPath() string               { return value.currentPath }
func (value HistorySupersession) CurrentSHA256() string             { return value.currentSHA256 }
func (value HistorySupersession) SupersededPath() string            { return value.supersededPath }
func (value HistorySupersession) SupersededSHA256() string          { return value.supersededSHA256 }

func bindHistorySupersessions(
	selected []postgresv1.SelectedMigration,
	documents []Document,
) (map[string][]HistorySupersession, error) {
	selectedByPath := make(map[string]postgresv1.SelectedMigration, len(selected))
	for _, migration := range selected {
		selectedByPath[migration.Path] = migration
	}
	byScope := map[string][]HistorySupersession{}
	versions := map[int64]bool{}
	for _, document := range documents {
		if len(document.superseded) == 0 {
			continue
		}
		current, exists := selectedByPath[document.path]
		if !exists {
			continue
		}
		if current.Owner != document.owner || current.Scope != document.scope ||
			versions[current.Version] {
			return nil, fmt.Errorf("migration history supersession binding is invalid")
		}
		versions[current.Version] = true
		for _, superseded := range document.superseded {
			byScope[current.Scope] = append(byScope[current.Scope], HistorySupersession{
				version: current.Version, owner: current.Owner, scope: current.Scope,
				origin: current.Origin, currentPath: current.Path,
				currentSHA256: current.SHA256, supersededPath: superseded.path,
				supersededSHA256: superseded.sha256,
			})
		}
	}
	for scope := range byScope {
		sort.Slice(byScope[scope], func(left, right int) bool {
			if byScope[scope][left].version != byScope[scope][right].version {
				return byScope[scope][left].version < byScope[scope][right].version
			}
			return byScope[scope][left].supersededPath <
				byScope[scope][right].supersededPath
		})
	}
	return byScope, nil
}
