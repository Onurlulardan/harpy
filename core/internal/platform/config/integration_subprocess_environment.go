//go:build integration

package config

import (
	"os"
	"strings"
)

func IntegrationSubprocessEnvironment(overrides []string) []string {
	environment := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "HARPY_") {
			environment = append(environment, entry)
		}
	}
	return append(environment, overrides...)
}
