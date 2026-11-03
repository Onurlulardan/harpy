package main

import "fmt"

func changedMigrations(base, current map[string]string) error {
	for path, before := range base {
		after, ok := current[path]
		if !ok {
			return fmt.Errorf("applied goose migration removed: %s", path)
		}
		if after != before {
			return fmt.Errorf("applied goose migration changed: %s", path)
		}
	}
	return nil
}
