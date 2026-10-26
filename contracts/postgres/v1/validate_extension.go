package postgresv1

import (
	"fmt"
	"strings"
)

func (manifest ExtensionManifest) Validate() error {
	origin := DatabaseOrigin(manifest.DatabaseOrigin)
	policy := manifest.ControlPathPolicy
	if manifest.ContractVersion != "postgres-extension-manifest.v1" || !origin.Valid() ||
		manifest.PostgresImageDigest != PostgresImageDigest ||
		manifest.PostgresEngineVersion != PostgresEngineVersion {
		return fmt.Errorf("extension manifest header is invalid")
	}
	if policy.Resolver != "pg_config_sharedir" || policy.RelativeDirectory != "extension" ||
		policy.AllowAbsolutePaths || policy.AllowSymlinks {
		return fmt.Errorf("extension control path policy is invalid")
	}
	if len(manifest.Extensions) < 2 || len(manifest.Extensions) > 3 {
		return fmt.Errorf("extension inventory is incomplete")
	}
	keys := make([]string, 0, len(manifest.Extensions))
	required := map[string]bool{"pg_trgm": false, "vector": false}
	for index, extension := range manifest.Extensions {
		if err := validateExtension(extension, origin); err != nil {
			return fmt.Errorf("extension %d: %w", index, err)
		}
		keys = append(keys, extension.Name)
		if _, exists := required[extension.Name]; exists {
			required[extension.Name] = true
		}
	}
	if err := validateOrderedUnique(keys, "extensions"); err != nil {
		return err
	}
	if !required["pg_trgm"] || !required["vector"] {
		return fmt.Errorf("required extension is missing")
	}
	return nil
}

func validateExtension(extension PostgresExtension, origin DatabaseOrigin) error {
	capability, valid := extensionCapability(extension.Name)
	if !valid || !versionPattern.MatchString(extension.Version) {
		return fmt.Errorf("name or version is invalid")
	}
	if extension.Name == "vector" && extension.Version != "0.8.3" {
		return fmt.Errorf("vector version is invalid")
	}
	if err := validateIdentifier(extension.Schema, "schema"); err != nil {
		return err
	}
	if err := validateExtensionOwner(extension.Owner, origin); err != nil {
		return err
	}
	if origin == DatabaseOriginFresh &&
		(extension.Schema != "platform_ext" || !extension.Install) {
		return fmt.Errorf("fresh extension placement is invalid")
	}
	if extension.ControlFile != extension.Name+".control" {
		return fmt.Errorf("control file is invalid")
	}
	if err := validateDigest(extension.ControlFileSHA256, "control_file_sha256"); err != nil {
		return err
	}
	if len(extension.Capabilities) != 1 || extension.Capabilities[0] != capability {
		return fmt.Errorf("extension capabilities are invalid")
	}
	return nil
}

func validateExtensionOwner(owner string, origin DatabaseOrigin) error {
	if err := validateIdentifier(owner, "extension owner"); err != nil {
		return err
	}
	if origin == DatabaseOriginFresh && owner != "harpy_extension_admin" {
		return fmt.Errorf("fresh extension owner is invalid")
	}
	if origin == DatabaseOriginV1Adopted && strings.HasPrefix(owner, "harpy_") &&
		owner != "harpy_extension_admin" && owner != "harpy_legacy_extension_owner" {
		return fmt.Errorf("adopted extension owner is invalid")
	}
	return nil
}

func extensionCapability(name string) (string, bool) {
	switch name {
	case "pg_trgm":
		return "trigram_search", true
	case "unaccent":
		return "unaccent_normalization", true
	case "vector":
		return "vector_similarity", true
	default:
		return "", false
	}
}
