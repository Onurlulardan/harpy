package installationcredential

import "errors"

var (
	ErrInvalid = errors.New("IAM installation credential command is invalid")
	ErrStorage = errors.New("IAM installation credential storage failed")
)
