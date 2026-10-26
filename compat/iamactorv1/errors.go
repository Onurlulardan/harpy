package iamactorv1

import "errors"

type Code string

const (
	CodeInvalid      Code = "IAM_V1_ACTOR_INVALID"
	CodeUnauthorized Code = "IAM_V1_ACTOR_UNAUTHORIZED"
	CodeExpired      Code = "IAM_V1_ACTOR_EXPIRED"
	CodeReplay       Code = "IAM_V1_ACTOR_REPLAY"
	CodeAuthority    Code = "IAM_V1_ACTOR_AUTHORITY_CHANGED"
	CodeRetired      Code = "IAM_V1_ACTOR_RETIRED"
	CodeUnavailable  Code = "IAM_V1_ACTOR_UNAVAILABLE"
)

type Error struct{ code Code }

func (failure Error) Error() string { return string(failure.code) }
func (failure Error) Code() Code    { return failure.code }

func newError(code Code) error {
	switch code {
	case CodeInvalid, CodeUnauthorized, CodeExpired, CodeReplay,
		CodeAuthority, CodeRetired, CodeUnavailable:
		return Error{code: code}
	default:
		return Error{code: CodeUnavailable}
	}
}

func Is(err error, code Code) bool {
	var failure Error
	return errors.As(err, &failure) && failure.code == code
}

type sqlStateError interface{ SQLState() string }

func classifyDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	var state sqlStateError
	if !errors.As(err, &state) {
		return newError(CodeUnavailable)
	}
	switch state.SQLState() {
	case "22023":
		return newError(CodeInvalid)
	case "42501":
		return newError(CodeUnauthorized)
	case "P5601", "P5602":
		return newError(CodeAuthority)
	case "P5603", "P5607":
		return newError(CodeRetired)
	case "P5604":
		return newError(CodeReplay)
	case "P5605":
		return newError(CodeExpired)
	default:
		return newError(CodeUnavailable)
	}
}
