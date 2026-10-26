package api

type Direction uint8

const (
	DirectionAscending Direction = iota + 1
	DirectionDescending
)

func ParseDirection(value string) (Direction, bool) {
	switch value {
	case "asc":
		return DirectionAscending, true
	case "desc":
		return DirectionDescending, true
	default:
		return 0, false
	}
}

func (direction Direction) String() string {
	switch direction {
	case DirectionAscending:
		return "asc"
	case DirectionDescending:
		return "desc"
	default:
		return ""
	}
}

func (direction Direction) valid() bool {
	return direction == DirectionAscending || direction == DirectionDescending
}

func (Direction) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
