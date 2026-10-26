package offering

type fieldSpec struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Column   string   `json:"column"`
	Folded   string   `json:"folded,omitempty"`
	Kind     string   `json:"kind"`
	Profile  string   `json:"profile,omitempty"`
	Filters  []string `json:"filters"`
	Sortable bool     `json:"sortable"`
	Cursor   int      `json:"cursor"`
	Search   bool     `json:"search"`
	Default  bool     `json:"default"`
}

var scalarFilters = []string{"eq", "ne", "in", "not_in"}
var orderedFilters = []string{
	"eq", "ne", "gt", "gte", "lt", "lte", "in", "not_in", "between",
}

var offeringFields = []fieldSpec{
	{fieldID, "id", "id", "", "uuid", "", scalarFilters, true, 36, false, true},
	{fieldOfferingID, "offering_id", "offering_id", "", "uuid", "", scalarFilters, true, 36, false, true},
	{fieldOfferingKey, "offering_key", "offering_key", "offering_key_folded", "search_text", "", []string{"eq", "ne", "contains", "starts_with", "ends_with", "in", "not_in"}, true, 127, true, true},
	{fieldName, "name", "name", "", "text", "", scalarFilters, true, 512, false, true},
	{fieldRevision, "revision", "revision", "", "integer", "", orderedFilters, true, 20, false, true},
	{fieldLifecycle, "lifecycle", "lifecycle", "", "enum", "lifecycle", scalarFilters, false, 0, false, true},
	{fieldVersion, "version", "version", "", "integer", "", orderedFilters, true, 20, false, true},
	{fieldBase, "is_base", "is_base", "", "boolean", "", scalarFilters, true, 5, false, true},
	{fieldPublic, "is_public", "is_public", "", "boolean", "", scalarFilters, true, 5, false, true},
	{fieldTrialDays, "trial_days", "trial_days", "", "integer", "days", orderedFilters, true, 20, false, true},
	{fieldGraceDays, "grace_days", "grace_days", "", "integer", "days", orderedFilters, true, 20, false, true},
	{fieldPriceCount, "expected_price_count", "expected_price_count", "", "integer", "price_count", orderedFilters, true, 20, false, true},
	{fieldGrantCount, "expected_grant_count", "expected_grant_count", "", "integer", "grant_count", orderedFilters, true, 20, false, true},
	{fieldEffectiveStart, "effective_start", "effective_start", "", "timestamp", "", orderedFilters, true, 27, false, true},
	{fieldEffectiveEnd, "effective_end", "effective_end", "", "timestamp", "", append(append([]string{}, orderedFilters...), "is_null", "is_not_null"), true, 27, false, true},
	{fieldPublicationState, "publication_state", "publication_state", "", "enum", "publication", scalarFilters, false, 0, false, true},
	{fieldPublishedAt, "published_at", "published_at", "", "timestamp", "", orderedFilters, true, 27, false, true},
	{fieldDescription, "description", "description", "", "text", "", scalarFilters, true, 2048, false, true},
	{fieldSortOrder, "sort_order", "sort_order", "", "integer", "", orderedFilters, true, 20, false, true},
	{fieldHighlightLabel, "highlight_label", "highlight_label", "", "text", "", scalarFilters, true, 50, false, true},
	{fieldColor, "color", "color", "", "text", "", scalarFilters, true, 20, false, true},
	{fieldBillingInterval, "billing_interval", "billing_interval", "", "text", "", scalarFilters, true, 7, false, true},
	{fieldCreatedAt, "created_at", "created_at", "", "timestamp", "", orderedFilters, true, 27, false, true},
	{fieldUpdatedAt, "updated_at", "updated_at", "", "timestamp", "", orderedFilters, true, 27, false, true},
	{fieldAvailable, "available", "available", "", "boolean", "", scalarFilters, true, 5, false, true},
}
