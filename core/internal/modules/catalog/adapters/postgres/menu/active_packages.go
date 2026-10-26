package menu

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func (storage *store) listActivePackages(ctx context.Context) (map[string]bool, error) {
	resultRows, err := storage.query(ctx, listActivePackagesSQL)
	if err != nil {
		return nil, unavailable()
	}
	defer resultRows.Close()
	result := map[string]bool{}
	for resultRows.Next() {
		var raw string
		if err = resultRows.Scan(&raw); err != nil {
			return nil, unavailable()
		}
		id, parseErr := packagecontract.ParsePackageID(raw)
		if parseErr != nil {
			return nil, unavailable()
		}
		result[id.String()] = true
	}
	if resultRows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}
