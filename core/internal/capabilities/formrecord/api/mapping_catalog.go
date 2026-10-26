package recordkernel

import (
	"crypto/sha256"
	"fmt"
	"slices"
)

const MappingCatalogContract = "forms-publication-mapping-catalog.v1"

func DigestMappings(values []PhysicalMapping) ([32]byte, error) {
	ordered := append([]PhysicalMapping(nil), values...)
	slices.SortFunc(ordered, comparePhysicalMapping)
	for index, mapping := range ordered {
		if !mapping.Valid() ||
			index > 0 && mapping.PhysicalName == ordered[index-1].PhysicalName {
			return [32]byte{}, fmt.Errorf("mapping catalog is invalid")
		}
	}
	document, err := canonicalObject(struct {
		Contract string            `json:"contract"`
		Mappings []PhysicalMapping `json:"mappings"`
	}{MappingCatalogContract, ordered}, MaximumCatalogBytes)
	if err != nil {
		return [32]byte{}, fmt.Errorf("mapping catalog is invalid")
	}
	return sha256.Sum256(document), nil
}

func comparePhysicalMapping(left, right PhysicalMapping) int {
	if result := compareText(left.PhysicalName, right.PhysicalName); result != 0 {
		return result
	}
	if result := compareText(left.Kind, right.Kind); result != 0 {
		return result
	}
	return slices.Compare(left.LogicalID[:], right.LogicalID[:])
}
