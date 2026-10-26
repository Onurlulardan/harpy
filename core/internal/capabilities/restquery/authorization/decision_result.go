package authorization

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
)

type DecisionResult struct {
	requirement [sha256.Size]byte
	allowed     bool
	validated   bool
}

func NewDecisionResult(
	requirement [sha256.Size]byte,
	allowed bool,
) (DecisionResult, error) {
	result := DecisionResult{
		requirement: requirement, allowed: allowed, validated: true,
	}
	if err := result.Validate(); err != nil {
		return DecisionResult{}, err
	}
	return result, nil
}

func (result DecisionResult) Validate() error {
	if !result.validated || result.requirement == [sha256.Size]byte{} {
		return fmt.Errorf("REST Query policy decision result is invalid")
	}
	return nil
}

func (result DecisionResult) RequirementSHA256() ([sha256.Size]byte, error) {
	if err := result.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return result.requirement, nil
}

func (result DecisionResult) Allowed() (bool, error) {
	if err := result.Validate(); err != nil {
		return false, err
	}
	return result.allowed, nil
}

func (DecisionResult) String() string   { return authorizationRedaction }
func (DecisionResult) GoString() string { return authorizationRedaction }
func (DecisionResult) LogValue() slog.Value {
	return authorizationLogValue()
}
func (result DecisionResult) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (DecisionResult) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }
