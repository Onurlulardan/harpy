package discount

type DurationKind string

const (
	Once      DurationKind = "once"
	Repeating DurationKind = "repeating"
	Forever   DurationKind = "forever"
)

type Duration struct {
	kind   DurationKind
	cycles int64
}

func OnceDuration() Duration    { return Duration{kind: Once, cycles: 1} }
func ForeverDuration() Duration { return Duration{kind: Forever} }

func RepeatingDuration(cycles int64) (Duration, error) {
	if cycles <= 0 {
		return Duration{}, ErrDuration
	}
	return Duration{kind: Repeating, cycles: cycles}, nil
}

func (duration Duration) Kind() DurationKind { return duration.kind }
func (duration Duration) Cycles() int64      { return duration.cycles }
func (duration Duration) Valid() bool        { return duration.valid() }

func (duration Duration) Applies(cycleOrdinal int64) (bool, error) {
	if cycleOrdinal <= 0 || !duration.valid() {
		return false, ErrDuration
	}
	switch duration.kind {
	case Once:
		return cycleOrdinal == 1, nil
	case Repeating:
		return cycleOrdinal <= duration.cycles, nil
	case Forever:
		return true, nil
	default:
		return false, ErrDuration
	}
}

func (duration Duration) valid() bool {
	return duration.kind == Once && duration.cycles == 1 ||
		duration.kind == Repeating && duration.cycles > 0 ||
		duration.kind == Forever && duration.cycles == 0
}
