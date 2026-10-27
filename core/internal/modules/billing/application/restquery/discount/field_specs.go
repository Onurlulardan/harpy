package discount

type fieldSpec struct {
	ID, Name, Column, Folded, Kind, Profile string
	Filters                                 []string
	Sortable                                bool
	Cursor                                  int
	Search, Default                         bool
}

var scalar = []string{"eq", "ne", "in", "not_in"}
var ordered = []string{"eq", "ne", "gt", "gte", "lt", "lte", "in", "not_in", "between"}
var nullableOrdered = append(append([]string{}, ordered...), "is_null", "is_not_null")
var nullableScalar = append(append([]string{}, scalar...), "is_null", "is_not_null")

var fields = []fieldSpec{
	{fieldID, "id", "id", "", "uuid", "", scalar, true, 36, false, true},
	{fieldDiscountID, "discount_id", "discount_id", "", "uuid", "", scalar, true, 36, false, true},
	{fieldCode, "code", "code", "code_folded", "search_text", "", []string{"eq", "ne", "contains", "starts_with", "ends_with", "in", "not_in"}, true, 64, true, true},
	{fieldRevision, "revision", "revision", "", "integer", "positive", ordered, true, 20, false, true},
	{fieldLifecycle, "lifecycle", "lifecycle", "", "enum", "lifecycle", scalar, false, 0, false, true},
	{fieldVersion, "version", "version", "", "integer", "positive", ordered, true, 20, false, true},
	{fieldKind, "discount_kind", "discount_kind", "", "enum", "kind", scalar, false, 0, false, true},
	{fieldFixedMinorUnits, "fixed_minor_units", "fixed_minor_units", "", "integer", "money", nullableOrdered, true, 20, false, true},
	{fieldCurrency, "currency", "currency", "", "text", "", nullableScalar, true, 3, false, true},
	{fieldCurrencyRegistry, "currency_registry_version", "currency_registry_version", "", "integer", "positive", nullableOrdered, true, 20, false, true},
	{fieldBasisPoints, "basis_points", "basis_points", "", "integer", "basis_points", nullableOrdered, true, 5, false, true},
	{fieldDuration, "duration", "duration", "", "enum", "duration", scalar, false, 0, false, true},
	{fieldRepeatingCycles, "repeating_cycles", "repeating_cycles", "", "integer", "positive", nullableOrdered, true, 20, false, true},
	{fieldPriority, "priority", "priority", "", "integer", "nonnegative", ordered, true, 11, false, true},
	{fieldGlobalUseLimit, "global_use_limit", "global_use_limit", "", "integer", "positive", nullableOrdered, true, 20, false, true},
	{fieldPerTenantUseLimit, "per_tenant_use_limit", "per_tenant_use_limit", "", "integer", "positive", nullableOrdered, true, 20, false, true},
	{fieldAllOfferings, "all_offerings", "all_offerings", "", "boolean", "", scalar, true, 5, false, true},
	{fieldExpectedOfferingCount, "expected_offering_count", "expected_offering_count", "", "integer", "count", ordered, true, 3, false, true},
	{fieldValidStart, "valid_start", "valid_start", "", "timestamp", "", ordered, true, 27, false, true},
	{fieldValidEnd, "valid_end", "valid_end", "", "timestamp", "", nullableOrdered, true, 27, false, true},
	{fieldPublishedAt, "published_at", "published_at", "", "timestamp", "", ordered, true, 27, false, true},
}
