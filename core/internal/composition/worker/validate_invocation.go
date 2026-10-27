package worker

import "fmt"

func validateInvocation(arguments []string) error {
	if len(arguments) != 0 {
		return fmt.Errorf("usage: core-worker")
	}
	return nil
}
