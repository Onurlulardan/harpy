package systemresourcequery

import (
	"errors"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

var ErrContract = errors.New("IAM system-resource query contract failed")

type Failure struct {
	code   restqueryv1.ErrorCode
	status int
	issues []runtimev1.Issue
}

func NormalizeError(err error) error {
	var queryError api.QueryError
	if !errors.As(err, &queryError) {
		return err
	}
	status, valid := queryError.HTTPStatus()
	if !valid {
		return ErrContract
	}
	return Failure{
		code: queryError.Code(), status: status, issues: queryError.Issues(),
	}
}

func (failure Failure) Error() string   { return string(failure.code) }
func (failure Failure) HTTPStatus() int { return failure.status }
func (failure Failure) Code() string    { return string(failure.code) }
func (failure Failure) Issues() []runtimev1.Issue {
	return append([]runtimev1.Issue(nil), failure.issues...)
}
