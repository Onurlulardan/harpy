package contractretirement

import "math"

type RemainingScopeCount struct{ value int64 }

func NewRemainingScopeCount(value int64) (RemainingScopeCount, error) {
	if value < 0 {
		return RemainingScopeCount{}, ErrInvalid
	}
	return RemainingScopeCount{value: value}, nil
}

func (count RemainingScopeCount) DatabaseValue() (int64, error) {
	if count.value < 0 {
		return 0, ErrInvalid
	}
	return count.value, nil
}

type Revision struct{ value int64 }

func NewRevision(value int64) (Revision, error) {
	if value < 1 || value > math.MaxInt64 {
		return Revision{}, ErrInvalid
	}
	return Revision{value: value}, nil
}

func (revision Revision) DatabaseValue() (int64, error) {
	if revision.value < 1 {
		return 0, ErrInvalid
	}
	return revision.value, nil
}
