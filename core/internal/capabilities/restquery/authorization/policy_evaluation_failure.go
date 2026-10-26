package authorization

type policyEvaluationFailure struct {
	cause error
}

func newPolicyEvaluationFailure(cause error) error {
	if cause == nil {
		return errPolicyEvaluationContract
	}
	return policyEvaluationFailure{cause: cause}
}

func (failure policyEvaluationFailure) Error() string {
	return errPolicyEvaluationContract.Error()
}

func (failure policyEvaluationFailure) Unwrap() error {
	return failure.cause
}

func (failure policyEvaluationFailure) Is(target error) bool {
	return target == errPolicyEvaluationContract
}
