package campaign

import "errors"

var (
	ErrInstant     = errors.New("billing campaign instant is invalid")
	ErrKey         = errors.New("billing campaign key is invalid")
	ErrPublication = errors.New("billing campaign publication is invalid")
	ErrVersion     = errors.New("billing campaign version is invalid")
	ErrDiscount    = errors.New("billing campaign discount is invalid")
)
