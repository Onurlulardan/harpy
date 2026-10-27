package reports

type fieldSpec struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Column     string   `json:"column"`
	Kind       string   `json:"kind"`
	Filters    []string `json:"filters"`
	Sortable   bool     `json:"sortable"`
	Cursor     int      `json:"cursor"`
	Default    bool     `json:"default"`
	Groupable  bool     `json:"groupable"`
	Aggregates []string `json:"aggregates"`
}

var scalarFilters = []string{"eq", "ne", "in", "not_in"}
var nullableScalarFilters = append(
	append([]string{}, scalarFilters...), "is_null", "is_not_null",
)
var orderedFilters = []string{
	"eq", "ne", "gt", "gte", "lt", "lte", "in", "not_in", "between",
}
var nullableOrderedFilters = append(
	append([]string{}, orderedFilters...), "is_null", "is_not_null",
)
