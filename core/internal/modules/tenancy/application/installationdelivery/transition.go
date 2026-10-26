package installationdelivery

import "time"

type Transition struct {
	State       State
	Revision    int64
	DatabaseNow time.Time
}

func (transition Transition) valid(expected State) bool {
	return transition.State == expected && transition.Revision > 0 && utc(transition.DatabaseNow)
}
