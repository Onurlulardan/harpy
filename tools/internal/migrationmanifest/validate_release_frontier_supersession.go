package migrationmanifest

import (
	"fmt"
	"path"
	"reflect"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func selectLineageMaximalRelease(
	releases []immutableRelease,
) (immutableRelease, error) {
	var descendants []immutableRelease
	for _, candidate := range releases {
		extendsAll := true
		for _, previous := range releases {
			if !releaseManifestExtends(previous.manifest, candidate.manifest) {
				extendsAll = false
				break
			}
		}
		if extendsAll {
			descendants = append(descendants, candidate)
		}
	}
	if len(descendants) != 1 {
		left, right := releases[0].digest, releases[len(releases)-1].digest
		return immutableRelease{}, fmt.Errorf(
			"immutable migration release lineage is ambiguous between %s and %s",
			left, right,
		)
	}
	return descendants[0], nil
}

func releaseManifestExtends(
	previous postgresv1.MigrationManifest,
	current postgresv1.MigrationManifest,
) bool {
	previousContract := previous
	previousContract.Entries = nil
	currentContract := current
	currentContract.Entries = nil
	if !reflect.DeepEqual(previousContract, currentContract) ||
		len(current.Entries) < len(previous.Entries) {
		return false
	}
	for index, entry := range previous.Entries {
		if !migrationEntryExtends(entry, current.Entries[index]) {
			return false
		}
	}
	return true
}

func migrationEntryExtends(
	previous postgresv1.MigrationEntry,
	current postgresv1.MigrationEntry,
) bool {
	if reflect.DeepEqual(previous, current) {
		return true
	}
	previousMetadata := previous
	previousMetadata.Variants = postgresv1.MigrationVariants{}
	currentMetadata := current
	currentMetadata.Variants = postgresv1.MigrationVariants{}
	if !reflect.DeepEqual(previousMetadata, currentMetadata) ||
		previous.Variants.Fresh == nil || previous.Variants.V1Adopted == nil ||
		current.Variants.Fresh == nil || current.Variants.V1Adopted == nil ||
		previous.Variants.Fresh.SHA256 != previous.Variants.V1Adopted.SHA256 ||
		current.Variants.Fresh.SHA256 != current.Variants.V1Adopted.SHA256 {
		return false
	}
	return migrationVariantSupersedes(
		previous.Variants.Fresh, current.Variants.Fresh, "fresh",
	) && migrationVariantSupersedes(
		previous.Variants.V1Adopted, current.Variants.V1Adopted, "v1_adopted",
	)
}

func migrationVariantSupersedes(
	previous *postgresv1.MigrationVariant,
	current *postgresv1.MigrationVariant,
	origin string,
) bool {
	previousRoot, previousLevel, previousValid :=
		migrationVariantSupersessionLevel(previous.Path, origin)
	currentRoot, currentLevel, currentValid :=
		migrationVariantSupersessionLevel(current.Path, origin)
	return previousValid && currentValid && previousRoot == currentRoot &&
		previousLevel < currentLevel &&
		path.Base(previous.Path) == path.Base(current.Path)
}

func migrationVariantSupersessionLevel(
	migrationPath, origin string,
) (string, int, bool) {
	originDirectory := path.Dir(migrationPath)
	if path.Base(originDirectory) != origin {
		return "", 0, false
	}
	parentDirectory := path.Dir(originDirectory)
	if path.Base(parentDirectory) == "supersessions" {
		return path.Dir(parentDirectory), 1, true
	}
	if path.Base(parentDirectory) == "v2" &&
		path.Base(path.Dir(parentDirectory)) == "supersessions" {
		return path.Dir(path.Dir(parentDirectory)), 2, true
	}
	if path.Base(parentDirectory) == "v3" &&
		path.Base(path.Dir(parentDirectory)) == "supersessions" {
		return path.Dir(path.Dir(parentDirectory)), 3, true
	}
	if path.Base(path.Dir(parentDirectory)) == "supersessions" {
		return "", 0, false
	}
	return parentDirectory, 0, true
}
