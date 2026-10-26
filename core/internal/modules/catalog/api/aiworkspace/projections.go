package aiworkspace

import (
	"time"

	"github.com/google/uuid"
)

type TargetReference struct {
	Kind string
	ID   uuid.UUID
}

func (target TargetReference) Valid() bool {
	return target.ID != uuid.Nil && (target.Kind == "FORM" || target.Kind == "WORKFLOW" ||
		target.Kind == "DASHBOARD" || target.Kind == "ENDPOINT" || target.Kind == "GROUP")
}

type Item struct {
	Path, Type, Slug, Title, Description, Icon, Snippet string
	Tags                                                []string
	UpdatedAt                                           time.Time
	Score                                               float64
	Requirements                                        []Requirement
}

type Field struct {
	Name, Label, Type string
	Required          bool
	Options           []string
}

type Detail struct {
	Path, Type, Slug, Title, Description, Icon string
	Tags                                       []string
	UpdatedAt                                  time.Time
	Target                                     TargetReference
	Fields                                     []Field
	RecordCount                                int64
	Requirements                               []Requirement
}

type Change struct {
	Path, Type, ChangeType, ChangedBy string
	ChangedAt                         time.Time
	Requirements                      []Requirement
}

type Version struct {
	Version                  int
	PublishedAt              time.Time
	PublishedBy, DiffSummary string
	IsActive                 bool
	Requirements             []Requirement
}

type History struct {
	Items        []Version
	Requirements []Requirement
}
