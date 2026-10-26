package backupreceipt

import "errors"

var (
	ErrInvalid     = errors.New("backup receipt is invalid")
	ErrNotVerified = errors.New("backup restore verification is missing")
	ErrBinding     = errors.New("backup receipt binding does not match")
	ErrExpired     = errors.New("backup receipt retention is insufficient")
	ErrLegalHold   = errors.New("backup receipt is under legal hold")
	ErrUnavailable = errors.New("backup receipt authority is unavailable")
)
