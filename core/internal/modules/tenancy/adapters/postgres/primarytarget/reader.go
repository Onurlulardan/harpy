package primarytarget

import (
	"context"

	targetapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/primarytarget"
)

type Row interface {
	Scan(...any) error
}

type QueryRow func(
	context.Context,
	string,
	...any,
) Row

type reader struct {
	queryRow QueryRow
}

func NewReader(queryRow QueryRow) (targetapp.Reader, error) {
	if queryRow == nil {
		return nil, errTargetUnavailable
	}
	return &reader{queryRow: queryRow}, nil
}

func (targetReader *reader) usable(ctx context.Context) bool {
	return targetReader != nil && targetReader.queryRow != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ targetapp.Reader = (*reader)(nil)
