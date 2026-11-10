package platformops

import "fmt"

func backupContract(stage string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database backup %s: %w", stage, err)
}
