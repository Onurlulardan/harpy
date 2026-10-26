package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func fixtureTextListType(testingT testing.TB) api.FieldType {
	testingT.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		testingT.Fatal(err)
	}
	fieldType, err := api.NewTextListFieldType(profile)
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}

func fixtureUUIDListType(testingT testing.TB) api.FieldType {
	testingT.Helper()
	fieldType, err := api.NewUUIDListFieldType()
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}

func fixtureJSONType(testingT testing.TB) api.FieldType {
	testingT.Helper()
	fieldType, err := api.NewJSONFieldType()
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}
