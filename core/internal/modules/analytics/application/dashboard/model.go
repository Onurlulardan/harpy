package dashboard

import (
	"time"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	"github.com/google/uuid"
)

type Widget struct {
	WidgetID, QueryVersionID          uuid.UUID
	Position                          uint16
	ParameterDocument, OutputDocument []byte
}

type Snapshot struct {
	DashboardID, VersionID                       uuid.UUID
	VersionNumber, RootRevision, VersionRevision uint64
	Lifecycle                                    analyticsdashboard.Lifecycle
	Visibility                                   analyticsdashboard.Visibility
	Published                                    bool
	Name                                         string
	LayoutDocument                               []byte
	LayoutSHA256, ManifestSHA256                 [32]byte
	Widgets                                      []Widget
	CreatedAt, UpdatedAt                         time.Time
}

type Page struct {
	Items []Snapshot
	After *uuid.UUID
}

type ACL struct {
	DashboardID, PrincipalID uuid.UUID
	PrincipalKind            string
	Level                    string
	Revision                 uint64
}
