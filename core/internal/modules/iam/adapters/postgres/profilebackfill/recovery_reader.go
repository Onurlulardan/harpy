package profilebackfill

import profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"

type RecoveryReader struct {
	value *store
}

func NewRecoveryReader() (*RecoveryReader, error) {
	value, err := newStore()
	if err != nil {
		return nil, err
	}
	return &RecoveryReader{value: value}, nil
}

func (reader *RecoveryReader) usable() bool {
	return reader != nil && reader.value != nil
}

func NewRecoveryStore() (profilebackfillapp.RecoveryStore, error) {
	return newStore()
}

var _ profilebackfillapp.RecoveryStore = (*store)(nil)
