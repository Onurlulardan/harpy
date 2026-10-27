package iam

import (
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func systemTransferExportActions(
	plan systemTransferExportPlan,
) ([]string, error) {
	set, err := plan.published.CatalogSet()
	if err != nil {
		return nil, systemtransfer.ErrExportCatalogPinStale
	}
	root, err := set.RootCatalog()
	if err != nil {
		return nil, systemtransfer.ErrExportCatalogPinStale
	}
	fields := make(map[uuid.UUID]struct{})
	for _, name := range plan.columns {
		if err = addSystemTransferExportField(root, fields, name); err != nil {
			return nil, err
		}
	}
	for _, sort := range plan.query.Sort() {
		if err = addSystemTransferExportField(root, fields, sort.Field()); err != nil {
			return nil, err
		}
	}
	if filter, present := plan.query.Where(); present {
		if err = addSystemTransferExportFilterFields(root, fields, filter); err != nil {
			return nil, err
		}
	}
	if _, present := plan.query.Search(); present {
		searchFields, searchErr := root.SearchFields()
		if searchErr != nil || len(searchFields) == 0 {
			return nil, systemtransfer.ErrExportQueryInvalid
		}
		for _, identifier := range searchFields {
			fields[identifier] = struct{}{}
		}
	}
	actions := map[string]bool{"export": true, "read": true}
	for identifier := range fields {
		classification, classificationErr := plan.published.Classification(identifier)
		action, actionErr := systemTransferExportClassificationAction(classification)
		if classificationErr != nil || actionErr != nil {
			return nil, systemtransfer.ErrExportAuthorizationRevoked
		}
		actions[action] = true
	}
	ordered := make([]string, 0, 4)
	for _, action := range []string{"export", "read", "read.confidential", "read.restricted"} {
		if actions[action] {
			ordered = append(ordered, action)
		}
	}
	return ordered, nil
}

func addSystemTransferExportField(
	catalog queryapp.ExportFieldCatalog,
	fields map[uuid.UUID]struct{},
	name string,
) error {
	field, found, err := catalog.FieldByPublicName(name)
	if err != nil || !found {
		return systemtransfer.ErrExportQueryInvalid
	}
	identifier, err := field.ID()
	if err != nil {
		return systemtransfer.ErrExportCatalogPinStale
	}
	fields[identifier] = struct{}{}
	return nil
}

func addSystemTransferExportFilterFields(
	catalog queryapp.ExportFieldCatalog,
	fields map[uuid.UUID]struct{},
	filter queryapp.ExportFilter,
) error {
	if predicate, present := filter.Predicate(); present {
		return addSystemTransferExportField(catalog, fields, predicate.Field())
	}
	children := filter.Children()
	if len(children) == 0 {
		return systemtransfer.ErrExportQueryInvalid
	}
	for _, child := range children {
		if err := addSystemTransferExportFilterFields(catalog, fields, child); err != nil {
			return err
		}
	}
	return nil
}
