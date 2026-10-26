package aggregatepage

import (
	"math"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestMetadataLoadsExactCompilerDeclarations(t *testing.T) {
	metadata, err := loadMetadata(fixtureAggregateStatement(t))
	if err != nil {
		t.Fatal(err)
	}
	if metadata.limit != 2 || len(metadata.groups) != 1 || len(metadata.measures) != 2 ||
		metadata.groups[0].key != "status" || metadata.measures[0].key != "row_count" ||
		metadata.measures[1].key != "revenue" {
		t.Fatalf("metadata declaration order changed: %#v", metadata)
	}
}

func TestMetadataRejectsDuplicateAndInvalidMeasureDeclarations(t *testing.T) {
	integerProfile, _ := api.NewIntegerProfile(0, math.MaxInt64)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	booleanType, _ := api.NewBooleanFieldType()
	for name, value := range map[string]metadata{
		"zero": {},
		"duplicate group": {
			groups: []column{{fieldType: integerType, key: "same"},
				{fieldType: integerType, key: "same"}},
			measures: []column{}, limit: 1,
		},
		"duplicate measure": {
			groups: []column{}, measures: []column{
				{fieldType: integerType, key: "same"},
				{fieldType: integerType, key: "same"},
			}, limit: 1,
		},
		"boolean measure": {
			groups:   []column{},
			measures: []column{{fieldType: booleanType, key: "flag"}}, limit: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if validateMetadata(value) != ErrContract {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
}
