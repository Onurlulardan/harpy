package accesstoken

import (
	"io"

	"github.com/google/uuid"
)

func GenerateJTI(random io.Reader) (uuid.UUID, error) {
	if random == nil {
		return uuid.Nil, ErrInvalidClaims
	}
	identifier, err := uuid.NewRandomFromReader(random)
	if err != nil || !validV4UUID(identifier) {
		return uuid.Nil, ErrInvalidClaims
	}
	return identifier, nil
}
