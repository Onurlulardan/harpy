package menu

import (
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/google/uuid"
)

const (
	operationGet     = "catalog.menu.get"
	operationUpdate  = "catalog.menu.update"
	permissionRead   = "catalog.read"
	permissionManage = "catalog.menu.manage"
)

type Snapshot struct {
	VersionID uuid.UUID
	Number    uint64
	Revision  uint64
	Items     []menudomain.Item
}

type UpdateRequest struct {
	ItemID           uuid.UUID
	Label            *string
	Order            *int
	Visible          *bool
	ExpectedRevision uint64
	IdempotencyKey   string
}

type ReplaceItem struct {
	ItemID, Kind, Label string
	ParentID, TargetID  *string
	Order               int
	Visible             bool
}

type ReplaceRequest struct {
	ExpectedRevision uint64
	Items            []ReplaceItem
	IdempotencyKey   string
}
