package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func fixtureField(
	testingT testing.TB,
	id string,
	publicName string,
	columnName string,
	fieldType api.FieldType,
	foldedName string,
	accentName string,
	operators []api.Operator,
	sortable bool,
	searchable bool,
) api.FieldDefinition {
	testingT.Helper()
	column := fixtureToken(testingT, columnName)
	var folded, accent *sqlidentifier.Token
	if foldedName != "" {
		token := fixtureToken(testingT, foldedName)
		folded = &token
	}
	if accentName != "" {
		token := fixtureToken(testingT, accentName)
		accent = &token
	}
	storage, err := api.NewFieldStorage(column, folded, accent)
	if err != nil {
		testingT.Fatal(err)
	}
	filters := make([]api.FilterCapabilityInput, len(operators))
	for index, operator := range operators {
		filters[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	cursorBytes := 0
	if sortable {
		cursorBytes = 256
	}
	capabilities, err := api.NewFieldCapabilities(fieldType, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: sortable,
		CursorCanonicalBytes: cursorBytes, Searchable: searchable,
		Aggregates: []api.AggregateFunction{},
	})
	if err != nil {
		testingT.Fatal(err)
	}
	field, err := api.NewFieldDefinition(
		id, publicName, fieldType, storage, capabilities,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return field
}

func fixtureFoldedTextType(testingT testing.TB, accent bool) api.FieldType {
	testingT.Helper()
	accentProfile := api.TextAccentDisabled
	if accent {
		accentProfile = api.TextAccentInsensitiveV1
	}
	profile, err := api.NewTextProfile(api.TextCasefoldUnicodeDefaultV1, accentProfile)
	if err != nil {
		testingT.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}

func fixtureUUIDType(testingT testing.TB) api.FieldType {
	testingT.Helper()
	fieldType, err := api.NewUUIDFieldType()
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}

func fixtureDecimalType(testingT testing.TB, precision, scale int) api.FieldType {
	testingT.Helper()
	profile, err := api.NewDecimalProfile(precision, scale)
	if err != nil {
		testingT.Fatal(err)
	}
	fieldType, err := api.NewDecimalFieldType(profile)
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}

func fixtureOrderedEnumType(testingT testing.TB) api.FieldType {
	testingT.Helper()
	closedOrdinal, openOrdinal := 20, 10
	closed, closedErr := api.NewEnumValue("CLOSED", &closedOrdinal)
	open, openErr := api.NewEnumValue("OPEN", &openOrdinal)
	profile, profileErr := api.NewEnumProfile(
		"eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee", []api.EnumValue{closed, open},
	)
	if closedErr != nil || openErr != nil || profileErr != nil {
		testingT.Fatal("enum fixture is invalid")
	}
	fieldType, err := api.NewEnumFieldType(profile)
	if err != nil {
		testingT.Fatal(err)
	}
	return fieldType
}
