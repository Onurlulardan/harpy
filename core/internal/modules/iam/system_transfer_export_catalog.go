package iam

import (
	"context"
	"crypto/subtle"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	querycatalog "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferExportPlan struct {
	published querycatalog.Published
	request   queryapp.Request
	query     queryapp.ExportQuery
	columns   []string
	kinds     []queryapp.ExportLogicalType
}

func loadSystemTransferExportPlan(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.ExportAttempt,
) (systemTransferExportPlan, error) {
	if ctx == nil || ctx.Err() != nil || attempt.Validate() != nil {
		return systemTransferExportPlan{}, systemtransfer.ErrExportCatalogPinStale
	}
	kind, err := systemTransferExportResourceKind(attempt.Resource)
	if err != nil {
		return systemTransferExportPlan{}, err
	}
	reader, err := iampostgres.BindProfileReader(scope)
	if err != nil {
		return systemTransferExportPlan{}, systemtransfer.ErrExportUnavailable
	}
	schema, err := reader.HistoricalSchema(ctx, attempt.CatalogVersionID)
	if err != nil || schema.ID() != attempt.CatalogVersionID || schema.Resource() != kind {
		return systemTransferExportPlan{}, systemtransfer.ErrExportCatalogPinStale
	}
	custom, err := profileapp.NewRetainedQueryCatalog(schema)
	if err != nil {
		return systemTransferExportPlan{}, systemtransfer.ErrExportCatalogPinStale
	}
	published, err := querycatalog.Build(custom)
	if err != nil || !systemTransferExportPinMatches(published, attempt) {
		return systemTransferExportPlan{}, systemtransfer.ErrExportCatalogPinStale
	}
	request, err := queryapp.DecodePOST(attempt.QueryDocument)
	if err != nil || queryapp.ValidateCatalogRequest(published, request) != nil {
		return systemTransferExportPlan{}, systemtransfer.ErrExportQueryInvalid
	}
	query, err := request.Value()
	if err != nil || query.Mode() != queryapp.ExportQueryModeRecord ||
		len(query.Includes()) != 0 || query.Count() != queryapp.ExportCountModeNone {
		return systemTransferExportPlan{}, systemtransfer.ErrExportQueryInvalid
	}
	if _, present := query.Page().After(); present {
		return systemTransferExportPlan{}, systemtransfer.ErrExportQueryInvalid
	}
	columns, kinds, err := systemTransferExportColumns(published, query.Select())
	if err != nil {
		return systemTransferExportPlan{}, err
	}
	return systemTransferExportPlan{published: published, request: request,
		query: query, columns: columns, kinds: kinds}, nil
}

func systemTransferExportResourceKind(
	resource publicresource.Resource,
) (profileapp.ResourceKind, error) {
	switch resource {
	case publicresource.Users:
		return profileapp.ResourceUser, nil
	case publicresource.Roles:
		return profileapp.ResourceRole, nil
	case publicresource.Organizations:
		return profileapp.ResourceOrganization, nil
	default:
		return "", systemtransfer.ErrExportQueryInvalid
	}
}

func systemTransferExportPinMatches(
	published querycatalog.Published,
	attempt systemtransfer.ExportAttempt,
) bool {
	pin, err := systemTransferPublishedPin(published)
	return err == nil && pin.versionID == attempt.CatalogVersionID &&
		subtle.ConstantTimeCompare(pin.sha256[:], attempt.CatalogSHA256[:]) == 1
}

func systemTransferExportColumns(
	published querycatalog.Published,
	selected []string,
) ([]string, []queryapp.ExportLogicalType, error) {
	set, err := published.CatalogSet()
	if err != nil {
		return nil, nil, systemtransfer.ErrExportCatalogPinStale
	}
	root, err := set.RootCatalog()
	if err != nil {
		return nil, nil, systemtransfer.ErrExportCatalogPinStale
	}
	if len(selected) == 0 {
		defaults, defaultErr := root.DefaultSelect()
		if defaultErr != nil || len(defaults) == 0 {
			return nil, nil, systemtransfer.ErrExportQueryInvalid
		}
		selected = make([]string, len(defaults))
		for index, identifier := range defaults {
			field, found, fieldErr := root.Field(identifier)
			if fieldErr != nil || !found {
				return nil, nil, systemtransfer.ErrExportCatalogPinStale
			}
			name, nameErr := field.PublicName()
			if nameErr != nil {
				return nil, nil, systemtransfer.ErrExportCatalogPinStale
			}
			selected[index] = name
		}
	}
	kinds := make([]queryapp.ExportLogicalType, len(selected))
	for index, name := range selected {
		field, found, fieldErr := root.FieldByPublicName(name)
		if fieldErr != nil || !found {
			return nil, nil, systemtransfer.ErrExportQueryInvalid
		}
		fieldType, typeErr := field.Type()
		kinds[index], err = fieldType.Kind()
		if typeErr != nil || err != nil {
			return nil, nil, systemtransfer.ErrExportCatalogPinStale
		}
	}
	return append([]string(nil), selected...), kinds, nil
}
