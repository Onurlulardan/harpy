package api

type AggregateReference uint8

const (
	AggregateReferenceGroup AggregateReference = iota + 1
	AggregateReferenceMeasure
)

func (reference AggregateReference) String() string {
	switch reference {
	case AggregateReferenceGroup:
		return "group"
	case AggregateReferenceMeasure:
		return "measure"
	default:
		return ""
	}
}

func (reference AggregateReference) valid() bool {
	return reference == AggregateReferenceGroup || reference == AggregateReferenceMeasure
}

func (AggregateReference) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
