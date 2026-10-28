package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type profileEnvironment map[string]string

func addProfileEnvironmentArtifacts(set *artifactSet, _ time.Time) error {
	profiles := []struct {
		name   string
		values profileEnvironment
	}{
		{"api.env", apiEnvironment(set)},
		{"worker.env", workerEnvironment(set)},
		{"migrator.env", migratorEnvironment(set)},
		{"schema-executor.env", schemaExecutorEnvironment(set)},
		{"operator.env", operatorEnvironment(set)},
		{"tenant-migration-executor.env", tenantMigrationExecutorEnvironment(set)},
	}
	if set.options.profile == profileForgeGuard {
		profiles = append(profiles,
			struct {
				name   string
				values profileEnvironment
			}{"delivery-operator.env", deliveryOperatorEnvironment(set)},
		)
	}
	for _, profile := range profiles {
		content, err := renderProfileEnvironment(profile.values)
		if err != nil {
			return fmt.Errorf("render %s: %w", profile.name, err)
		}
		if err := set.add(profile.name, content); err != nil {
			return err
		}
	}
	return nil
}

func renderProfileEnvironment(values profileEnvironment) ([]byte, error) {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if !validEnvironmentKey(key) || value == "" ||
			strings.ContainsAny(value, "\x00\r\n") {
			return nil, fmt.Errorf("invalid environment entry %q", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var output strings.Builder
	for _, key := range keys {
		output.WriteString(key)
		output.WriteByte('=')
		output.WriteString(values[key])
		output.WriteByte('\n')
	}
	return []byte(output.String()), nil
}

func validEnvironmentKey(value string) bool {
	if value == "" || (value[0] < 'A' || value[0] > 'Z') {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func baseRuntimeEnvironment(dsn string) profileEnvironment {
	set := newArtifactSet()
	return baseRuntimeEnvironmentFor(set, dsn)
}

func baseRuntimeEnvironmentFor(
	set *artifactSet,
	dsn string,
) profileEnvironment {
	environment := "development"
	if set.options.profile == profileForgeGuard {
		environment = "production"
	}
	values := profileEnvironment{
		"HARPY_ENVIRONMENT":            environment,
		"HARPY_POSTGRES_ENDPOINT_KIND": "direct_postgres",
		"HARPY_POSTGRES_DSN_FILE":      dsn,
	}
	if set.options.profile == profileForgeGuard {
		values["HARPY_OTLP_ENDPOINT"] = "https://telemetry:4318"
		values["HARPY_OTLP_INSECURE"] = "false"
	}
	return values
}
