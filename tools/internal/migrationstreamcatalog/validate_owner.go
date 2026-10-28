package migrationstreamcatalog

import (
	"fmt"
	"io/fs"
	"strings"
)

func validateOwner(document ownerDocument, expectedName string) error {
	if document.Owner != expectedName {
		return fmt.Errorf(
			"migration stream owner %q differs from fragment name %q",
			document.Owner, expectedName,
		)
	}
	if document.Streams == nil {
		return fmt.Errorf("migration stream owner %q streams must not be null", document.Owner)
	}
	if len(document.Streams) == 0 {
		return fmt.Errorf("migration stream owner %q streams must not be empty", document.Owner)
	}
	seen := make(map[string]bool, len(document.Streams))
	previousVersion := int64(0)
	for index, stream := range document.Streams {
		if err := validateStream(stream, document.Owner); err != nil {
			return err
		}
		identity := fmt.Sprintf(
			"%s|%s|%s|%d|%d|%s|%d|%d", stream.Owner, stream.Directory,
			stream.Scope, stream.VersionFrom, stream.VersionTo, stream.Phase,
			stream.TargetDatabaseGeneration, stream.NextMinimumRuntimeGeneration,
		)
		if seen[identity] || index > 0 && previousVersion == stream.VersionFrom {
			return fmt.Errorf("duplicate migration stream for owner %q", document.Owner)
		}
		if index > 0 && previousVersion > stream.VersionFrom {
			return fmt.Errorf("migration streams for owner %q must be sorted", document.Owner)
		}
		seen[identity] = true
		previousVersion = stream.VersionFrom
	}
	return nil
}

func validateStream(stream Stream, expectedOwner string) error {
	if stream.Owner != expectedOwner {
		return fmt.Errorf("migration stream owner %q differs from fragment owner", stream.Owner)
	}
	if !fs.ValidPath(stream.Directory) || strings.ContainsAny(stream.Directory, `\:`) {
		return fmt.Errorf("migration stream directory %q is invalid", stream.Directory)
	}
	if stream.Scope != "platform" && stream.Scope != "tenant" {
		return fmt.Errorf("migration stream scope %q is invalid", stream.Scope)
	}
	if stream.VersionFrom < 1 || stream.VersionFrom > stream.VersionTo {
		return fmt.Errorf("migration stream version range is invalid")
	}
	if !validPhase(stream.Phase) {
		return fmt.Errorf("migration stream phase %q is invalid", stream.Phase)
	}
	if stream.TargetDatabaseGeneration < 1 ||
		stream.NextMinimumRuntimeGeneration < 1 {
		return fmt.Errorf("migration stream generation is invalid")
	}
	return nil
}

func validPhase(phase string) bool {
	switch phase {
	case "foundation", "expand", "backfill", "switch", "contract":
		return true
	default:
		return false
	}
}
