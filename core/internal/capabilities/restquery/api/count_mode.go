package api

type CountMode uint8

const (
	CountModeNone CountMode = iota + 1
	CountModeExact
)

func ParseCountMode(value string) (CountMode, bool) {
	switch value {
	case "none":
		return CountModeNone, true
	case "exact":
		return CountModeExact, true
	default:
		return 0, false
	}
}

func (mode CountMode) String() string {
	switch mode {
	case CountModeNone:
		return "none"
	case CountModeExact:
		return "exact"
	default:
		return ""
	}
}

func (mode CountMode) valid() bool {
	return mode == CountModeNone || mode == CountModeExact
}

func (CountMode) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
