package sqlidentifier

import "errors"

func (Token) MarshalJSON() ([]byte, error) {
	return nil, errors.New("SQL identifier token cannot be encoded as JSON")
}
