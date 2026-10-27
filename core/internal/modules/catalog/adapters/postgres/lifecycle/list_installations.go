package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_installations.sql
var listInstallationsSQL string

func (storage *store) ListInstallations(
	ctx context.Context, request catalogapp.PageRequest,
) (catalogapp.InstallationPage, error) {
	if request.Limit < 1 || request.Limit > 100 {
		return catalogapp.InstallationPage{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return catalogapp.InstallationPage{}, unavailable()
	}
	resultRows, err := storage.query(ctx, listInstallationsSQL, request.After, request.Limit+1)
	if err != nil {
		return catalogapp.InstallationPage{}, unavailable()
	}
	defer resultRows.Close()
	page := catalogapp.InstallationPage{Items: make([]catalogdomain.Installation, 0, request.Limit)}
	for resultRows.Next() {
		value, scanErr := scanInstallation(resultRows)
		if scanErr != nil {
			return catalogapp.InstallationPage{}, unavailable()
		}
		if len(page.Items) == request.Limit {
			next, parseErr := uuid.Parse(page.Items[len(page.Items)-1].ID.String())
			if parseErr != nil {
				return catalogapp.InstallationPage{}, unavailable()
			}
			page.Next = next
			break
		}
		page.Items = append(page.Items, value)
	}
	if resultRows.Err() != nil {
		return catalogapp.InstallationPage{}, unavailable()
	}
	return page, nil
}
