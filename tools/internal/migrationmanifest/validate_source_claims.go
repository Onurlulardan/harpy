package migrationmanifest

import (
	"fmt"
	"strings"
)

func validateSourceClaims(inventory StorageInventory, groups []SourceGroup) error {
	if len(inventory.streams) == 0 || len(groups) != len(inventory.streams) {
		return fmt.Errorf("migration stream set differs from source groups")
	}
	used := make(map[int]bool, len(groups))
	for _, group := range groups {
		streamIndex := matchingStream(inventory.streams, group)
		if streamIndex < 0 || used[streamIndex] {
			return fmt.Errorf("migration source group differs from migration stream contract")
		}
		used[streamIndex] = true
		stream := inventory.streams[streamIndex]
		prefix := group.Directory + "/"
		for _, document := range group.Documents {
			identity, err := parseDocumentIdentity(document.Path())
			if err != nil || document.Owner() != group.Owner ||
				!strings.HasPrefix(document.Path(), prefix) ||
				identity.version < stream.VersionFrom || identity.version > stream.VersionTo {
				return fmt.Errorf("migration source escapes its authoritative stream")
			}
		}
	}
	return nil
}

func matchingStream(streams []MigrationStream, group SourceGroup) int {
	for index, stream := range streams {
		if stream.Owner == group.Owner && stream.Directory == group.Directory &&
			stream.Scope == group.Scope && stream.Phase == group.Phase &&
			stream.VersionFrom == group.VersionFrom && stream.VersionTo == group.VersionTo &&
			stream.TargetDatabaseGeneration == group.TargetDatabaseGeneration &&
			stream.NextMinimumRuntimeGeneration == group.NextMinimumRuntimeGeneration {
			return index
		}
	}
	return -1
}
