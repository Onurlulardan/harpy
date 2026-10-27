package profile

type Cardinality string

const (
	CardinalitySingle Cardinality = "single"
	CardinalityList   Cardinality = "list"
)

func (cardinality Cardinality) Valid() bool {
	return cardinality == CardinalitySingle || cardinality == CardinalityList
}

func validTypeCardinality(kind LogicalType, cardinality Cardinality) bool {
	if cardinality == CardinalitySingle {
		return true
	}
	return cardinality == CardinalityList &&
		(kind == LogicalString || kind == LogicalUUID)
}
