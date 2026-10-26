package migrator

import (
	"errors"
	"fmt"
)

var errDeployRuntime = errors.New("deploy runtime failed")

func deployRuntimeError(cause error) error {
	if cause == nil {
		return errDeployRuntime
	}
	return fmt.Errorf("%w: %w", errDeployRuntime, cause)
}
