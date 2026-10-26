package environmentoperator

import (
	"errors"
	"fmt"
)

func recoverysigningZero() RecoverySigningResult {
	return RecoverySigningResult{}
}

func (result RecoverySigningResult) IsValid() bool {
	return result.Generation == 1 && len(result.TrustBundleSHA256) == 64
}

func classifyDependency(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalid):
		return ErrInvalid
	case errors.Is(err, ErrDependencyMismatch):
		return ErrDependencyMismatch
	case errors.Is(err, ErrFenced):
		return ErrFenced
	case errors.Is(err, ErrFailed):
		return ErrFailed
	default:
		return fmt.Errorf("%w: %w", ErrResumable, err)
	}
}

func (result InitializeResult) IsValid() bool {
	return (result.Outcome == OutcomeCompleted || result.Outcome == OutcomeNoop) &&
		result.OperationID != [16]byte{} && result.TenantID != [16]byte{} &&
		result.FirstUserID != [16]byte{}
}
