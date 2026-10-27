package recoveryorchestration

import (
	"fmt"
)

func restoreLocators(
	controlDSN, scratchDSN string,
	digest func(string) (string, error),
) (string, string, error) {
	if digest == nil {
		return "", "", fmt.Errorf("restore database locators are invalid")
	}
	control, controlErr := digest(controlDSN)
	scratch, scratchErr := digest(scratchDSN)
	if controlErr != nil || scratchErr != nil || control == scratch {
		return "", "", fmt.Errorf("restore database locators are invalid")
	}
	return control, scratch, nil
}
