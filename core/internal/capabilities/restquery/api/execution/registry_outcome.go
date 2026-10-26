package execution

import "errors"

func mergePortOutcome(portErr, programErr error) error {
	if programErr == nil {
		return portErr
	}
	if portErr == nil || errors.Is(portErr, programErr) {
		if portErr != nil {
			return portErr
		}
		return programErr
	}
	return errors.Join(programErr, portErr)
}
