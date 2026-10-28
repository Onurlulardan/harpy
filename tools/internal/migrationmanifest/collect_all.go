package migrationmanifest

import "fmt"

func CollectAll(
	repositoryRoot string,
	inventory StorageInventory,
) ([]SourceGroup, error) {
	groups := make([]SourceGroup, len(inventory.streams))
	for index, stream := range inventory.streams {
		groups[index] = SourceGroup{
			Owner: stream.Owner, Scope: stream.Scope, Phase: stream.Phase,
			TargetDatabaseGeneration:     stream.TargetDatabaseGeneration,
			NextMinimumRuntimeGeneration: stream.NextMinimumRuntimeGeneration,
			VersionFrom:                  stream.VersionFrom,
			VersionTo:                    stream.VersionTo,
			Directory:                    stream.Directory,
		}
	}
	for _, claim := range inventory.claims {
		collected, err := Collect(repositoryRoot, SourceSpec{
			Owner: claim.Owner, Scope: claim.Scope, Phase: "foundation",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			VersionFrom: 1, VersionTo: 1,
			Directory: claim.Directory,
		})
		if err != nil {
			return nil, err
		}
		for _, document := range collected.Documents {
			identity, err := parseDocumentIdentity(document.Path())
			if err != nil {
				return nil, err
			}
			streamIndex := streamForDocument(inventory.streams, claim, identity.version)
			if streamIndex < 0 {
				return nil, fmt.Errorf(
					"migration source %q owner %q scope %q version %d has no authoritative stream",
					document.Path(), claim.Owner, claim.Scope, identity.version,
				)
			}
			groups[streamIndex].Documents = append(groups[streamIndex].Documents, document)
		}
	}
	for _, group := range groups {
		if len(group.Documents) == 0 {
			return nil, fmt.Errorf("migration stream has no source documents")
		}
	}
	return groups, nil
}

func streamForDocument(
	streams []MigrationStream,
	claim StorageClaim,
	version int64,
) int {
	matched := -1
	for index, stream := range streams {
		if stream.Owner == claim.Owner && stream.Directory == claim.Directory &&
			stream.Scope == claim.Scope && version >= stream.VersionFrom &&
			version <= stream.VersionTo {
			if matched >= 0 {
				return -1
			}
			matched = index
		}
	}
	return matched
}
