package billing

import (
	"math"

	"github.com/google/uuid"
)

func parseGlobalHTTPV7(value string) (uuid.UUID, error) {
	identifier, err := uuid.Parse(value)
	if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
		return uuid.Nil, errInvalidInput
	}
	return identifier, nil
}

func int32Value(value int) (int32, error) {
	if int64(value) < math.MinInt32 || int64(value) > math.MaxInt32 {
		return 0, errInvalidInput
	}
	return int32(value), nil
}
