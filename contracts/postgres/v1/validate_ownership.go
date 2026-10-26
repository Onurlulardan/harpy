package postgresv1

import (
	"fmt"
	pathpkg "path"
	"strings"
)

func (manifest OwnershipManifest) Validate() error {
	origin := DatabaseOrigin(manifest.DatabaseOrigin)
	if manifest.ContractVersion != "postgres-ownership-manifest.v1" ||
		!uuidPattern.MatchString(manifest.ManifestID) || !origin.Valid() {
		return fmt.Errorf("ownership manifest identity is invalid")
	}
	if err := validateDatabaseName(manifest.DatabaseName, "database_name"); err != nil {
		return err
	}
	executor := ownershipExecutor(origin)
	if manifest.OwnershipExecutor != executor || manifest.ObjectCount < 1 ||
		manifest.ObjectCount > maximumSafeInteger {
		return fmt.Errorf("ownership executor or object count is invalid")
	}
	switch manifest.Format {
	case "inline":
		if manifest.Objects == nil || manifest.Chunks != nil ||
			len(manifest.Objects) == 0 || len(manifest.Objects) > 4096 ||
			manifest.ObjectCount != int64(len(manifest.Objects)) {
			return fmt.Errorf("inline ownership inventory is invalid")
		}
		return ValidateOwnershipObjects(manifest.Objects, origin, manifest.DatabaseName)
	case "chunked":
		if manifest.Objects != nil || len(manifest.Chunks) == 0 ||
			len(manifest.Chunks) > MaximumOwnershipChunks ||
			manifest.ObjectCount > MaximumOwnershipChunks*100000 {
			return fmt.Errorf("chunked ownership inventory is invalid")
		}
		return validateChunkDescriptors(manifest.Chunks)
	default:
		return fmt.Errorf("ownership format is invalid")
	}
}

func (chunk OwnershipChunk) Validate() error {
	origin := DatabaseOrigin(chunk.DatabaseOrigin)
	if chunk.ContractVersion != "postgres-ownership-chunk.v1" ||
		!uuidPattern.MatchString(chunk.ManifestID) || !origin.Valid() {
		return fmt.Errorf("ownership chunk identity is invalid")
	}
	if err := validateDatabaseName(chunk.DatabaseName, "database_name"); err != nil {
		return err
	}
	if chunk.Ordinal < 0 || chunk.Ordinal >= MaximumOwnershipChunks ||
		chunk.Count < 1 || chunk.Count > MaximumOwnershipChunks || chunk.Ordinal >= chunk.Count ||
		chunk.Objects == nil || len(chunk.Objects) == 0 || len(chunk.Objects) > 100000 ||
		chunk.ObjectCount < 1 || chunk.ObjectCount > maximumSafeInteger ||
		chunk.ObjectCount != int64(len(chunk.Objects)) {
		return fmt.Errorf("ownership chunk bounds are invalid")
	}
	return ValidateOwnershipObjects(chunk.Objects, origin, chunk.DatabaseName)
}

func validateChunkDescriptors(chunks []OwnershipChunkDescriptor) error {
	for index, chunk := range chunks {
		if chunk.Ordinal != index || chunk.Count != len(chunks) ||
			chunk.LengthBytes < 2 || chunk.LengthBytes > MaximumOwnershipChunkBytes {
			return fmt.Errorf("ownership chunk descriptor %d bounds are invalid", index)
		}
		if err := validateDigest(chunk.SHA256, "chunk sha256"); err != nil {
			return err
		}
		if err := validateArtifactPath(chunk.Path, "chunk path"); err != nil {
			return err
		}
		if pathpkg.Base(chunk.Path) != chunk.SHA256+".json" {
			return fmt.Errorf("ownership chunk path is not content addressed")
		}
	}
	return nil
}

func ValidateOwnershipObjects(
	objects []OwnershipObject,
	origin DatabaseOrigin,
	databaseName string,
) error {
	if !origin.Valid() {
		return fmt.Errorf("ownership object origin is invalid")
	}
	expectedExecutor := ownershipExecutor(origin)
	keys := make([]string, 0, len(objects))
	for index, object := range objects {
		if err := validateOwnershipObject(object, expectedExecutor); err != nil {
			return fmt.Errorf("ownership object %d: %w", index, err)
		}
		if object.Object.DatabaseName != databaseName {
			return fmt.Errorf("ownership object %d targets another database", index)
		}
		keys = append(keys, OwnershipObjectKey(object))
	}
	return validateOrderedUnique(keys, "ownership objects")
}

func validateOwnershipObject(object OwnershipObject, expectedExecutor string) error {
	if err := object.Object.Validate(); err != nil {
		return err
	}
	if err := validateIdentifier(object.Owner, "owner"); err != nil {
		return err
	}
	switch object.Classification {
	case "first_party":
		if object.Owner != "harpy_owner" ||
			(object.Executor != expectedExecutor && object.Executor != "normal_migrator") {
			return fmt.Errorf("first-party ownership decision is invalid")
		}
		if object.Object.Kind == "database" && object.Executor != expectedExecutor {
			return fmt.Errorf("database ownership executor is invalid")
		}
	case "extension":
		if object.Executor != "extension_admin" && object.Executor != expectedExecutor {
			return fmt.Errorf("extension ownership executor is invalid")
		}
		if strings.HasPrefix(object.Owner, "harpy_") &&
			object.Owner != "harpy_extension_admin" &&
			object.Owner != "harpy_legacy_extension_owner" {
			return fmt.Errorf("extension owner is invalid")
		}
	default:
		return fmt.Errorf("ownership classification is invalid")
	}
	return nil
}

func OwnershipObjectKey(object OwnershipObject) string {
	return strings.Join([]string{
		CatalogObjectKey(object.Object), object.Owner, object.Classification, object.Executor,
	}, "|")
}

func ownershipExecutor(origin DatabaseOrigin) string {
	if origin == DatabaseOriginV1Adopted {
		return "external_root_provisioner"
	}
	return "cluster_bootstrap"
}
