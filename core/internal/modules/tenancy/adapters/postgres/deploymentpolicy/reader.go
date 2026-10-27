package deploymentpolicy

import (
	"context"

	policyapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/deploymentpolicy"
)

type Row interface {
	Scan(...any) error
}

type QueryRow func(
	context.Context,
	string,
	...any,
) Row

type policyReader struct {
	queryRow QueryRow
}

func NewPolicyReader(queryRow QueryRow) (policyapp.PolicyReader, error) {
	if queryRow == nil {
		return nil, errPolicyUnavailable
	}
	return &policyReader{queryRow: queryRow}, nil
}

func (reader *policyReader) usable(ctx context.Context) bool {
	return reader != nil && reader.queryRow != nil &&
		ctx != nil && ctx.Err() == nil
}

var _ policyapp.PolicyReader = (*policyReader)(nil)
