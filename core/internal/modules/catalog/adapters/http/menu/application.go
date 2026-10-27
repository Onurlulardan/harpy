package menuhttp

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	"github.com/google/uuid"
)

type Application interface {
	GetMenu(context.Context, cataloghttp.Access) (cataloghttp.Menu, error)
	GetActiveMenuConfig(context.Context, cataloghttp.Access) (*cataloghttp.MenuConfig, error)
	CreateMenuConfig(context.Context, cataloghttp.Access, cataloghttp.MenuConfigWriteInput) (
		cataloghttp.MenuConfig, error)
	UpdateMenuConfig(context.Context, cataloghttp.Access, uuid.UUID,
		cataloghttp.MenuConfigWriteInput) (cataloghttp.MenuConfig, error)
	UpdateMenuOverride(context.Context, cataloghttp.Access, cataloghttp.MenuOverrideInput) (
		cataloghttp.MenuOverrideReceipt, error)
	ReplaceMenu(context.Context, cataloghttp.Access, cataloghttp.MenuReplaceInput) (
		cataloghttp.Menu, error)
}
