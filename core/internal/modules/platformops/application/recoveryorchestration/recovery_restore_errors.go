package recoveryorchestration

import (
	"errors"
	"fmt"
)

type databaseRestoreErrorClass uint8

const (
	restoreInvalid databaseRestoreErrorClass = iota + 1
	restoreDependency
	restoreResumable
	restoreTerminal
)

type RestoreFailureKind uint8

const (
	RestoreFailureInvalid RestoreFailureKind = iota + 1
	RestoreFailureDependency
	RestoreFailureResumable
	RestoreFailureTerminal
)

type databaseRestoreError struct {
	class databaseRestoreErrorClass
	cause error
}

func (failure *databaseRestoreError) Error() string {
	if failure == nil || failure.cause == nil {
		return "database restore failed"
	}
	return fmt.Sprintf("database restore failed: %v", failure.cause)
}

func (failure *databaseRestoreError) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.cause
}

func invalidRestore(err error) error {
	return &databaseRestoreError{class: restoreInvalid, cause: err}
}

func dependencyRestore(err error) error {
	return &databaseRestoreError{class: restoreDependency, cause: err}
}

func resumableRestore(err error) error {
	return &databaseRestoreError{class: restoreResumable, cause: err}
}

func terminalRestore(err error) error {
	return &databaseRestoreError{class: restoreTerminal, cause: err}
}

func restoreErrorClass(err error) databaseRestoreErrorClass {
	var failure *databaseRestoreError
	if errors.As(err, &failure) {
		return failure.class
	}
	return restoreResumable
}

func ClassifyDatabaseRestoreError(err error) RestoreFailureKind {
	switch restoreErrorClass(err) {
	case restoreInvalid:
		return RestoreFailureInvalid
	case restoreDependency:
		return RestoreFailureDependency
	case restoreTerminal:
		return RestoreFailureTerminal
	default:
		return RestoreFailureResumable
	}
}
