package recoverysigning

import "errors"

var (
	ErrInvalid     = errors.New("recovery signing input is invalid")
	ErrInterrupted = errors.New("recovery signing publication was interrupted")
	ErrFenced      = errors.New("recovery signing directory is fenced")
)

type Result struct {
	Generation        int64
	TrustBundleSHA256 string
	Reused            bool
}

func (result Result) IsValid() bool {
	return result.Generation == 1 && len(result.TrustBundleSHA256) == 64
}
