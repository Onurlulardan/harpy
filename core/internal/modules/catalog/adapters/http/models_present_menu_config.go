package cataloghttp

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
)

func RepresentMenuConfig(value menu.ConfigDocument) MenuConfig {
	return MenuConfig{
		ID: value.ID.String(), Name: value.Name, Config: value.Config,
		IsActive: value.Active, CreatedBy: value.CreatedBy.String(),
		UpdatedBy: value.UpdatedBy.String(), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}
