package campaign

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

var fields = []fieldSpec{
	{fieldID, "id", "id", "", "uuid", "", scalar, true, 36, false, true},
	{fieldCampaignID, "campaign_id", "campaign_id", "", "uuid", "", scalar, true, 36, false, true},
	{fieldCampaignKey, "campaign_key", "campaign_key", "campaign_key_folded", "search_text", "", []string{"eq", "ne", "contains", "starts_with", "ends_with", "in", "not_in"}, true, 128, true, true},
	{fieldRevision, "revision", "revision", "", "integer", "positive", ordered, true, 20, false, true},
	{fieldLifecycle, "lifecycle", "lifecycle", "", "enum", "lifecycle", scalar, false, 0, false, true},
	{fieldVersion, "version", "version", "", "integer", "positive", ordered, true, 20, false, true},
	{fieldName, "name", "name", "", "text", "", scalar, true, 128, false, true},
	{fieldBannerText, "banner_text", "banner_text", "", "text", "", scalar, false, 0, false, true},
	{fieldPriority, "priority", "priority", "", "integer", "signed", ordered, true, 11, false, true},
	{fieldDiscountCount, "expected_discount_count", "expected_discount_count", "", "integer", "count", ordered, true, 3, false, true},
	{fieldStartsAt, "starts_at", "starts_at", "", "timestamp", "", ordered, true, 27, false, true},
	{fieldEndsAt, "ends_at", "ends_at", "", "timestamp", "", nullableOrdered, true, 27, false, true},
	{fieldPublicationState, "publication_state", "publication_state", "", "enum", "publication", scalar, false, 0, false, true},
	{fieldPublishedAt, "published_at", "published_at", "", "timestamp", "", ordered, true, 27, false, true},
}
