package cutoverv1

import "fmt"

func contractError(subject string) error {
	return fmt.Errorf("cutover %s is invalid", subject)
}
