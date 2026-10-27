package application

import (
	"encoding/json"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func encodeCatalog(input []string) ([]string, string, error) {
	fields := append([]string(nil), input...)
	for _, fieldID := range fields {
		if !stableUUIDText(fieldID) {
			return nil, "", layoutapi.ErrInvalid
		}
	}
	sort.Slice(fields, func(left, right int) bool {
		return fields[left] < fields[right]
	})
	for index := 1; index < len(fields); index++ {
		if fields[index] == fields[index-1] {
			return nil, "", layoutapi.ErrInvalid
		}
	}
	document, err := json.Marshal(fields)
	if err != nil {
		return nil, "", layoutapi.ErrInvalid
	}
	canonical, err := canonicaljsonv1.Canonicalize(document, maximumDefinitionBytes)
	if err != nil {
		return nil, "", layoutapi.ErrInvalid
	}
	return fields, string(canonical), nil
}

func definitionUsesCatalogSubset(
	definition layoutapi.Definition,
	catalog []string,
) bool {
	allowed := make(map[string]struct{}, len(catalog))
	for _, fieldID := range catalog {
		allowed[fieldID] = struct{}{}
	}
	for _, surface := range definition.Surfaces {
		fields := make(map[string]struct{}, len(catalog))
		for _, section := range surface.Sections {
			for _, field := range section.Fields {
				if _, exists := allowed[field.FieldID]; !exists {
					return false
				}
				if _, duplicate := fields[field.FieldID]; duplicate {
					return false
				}
				fields[field.FieldID] = struct{}{}
			}
		}
	}
	return true
}
