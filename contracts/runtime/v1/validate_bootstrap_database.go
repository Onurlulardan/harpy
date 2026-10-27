package runtimev1

import (
	"fmt"
	"strings"
)

func validateBootstrapAuthorityDatabase(
	database BootstrapAuthorityDatabase,
	origin DatabaseOrigin,
) error {
	if !databaseNamePattern.MatchString(database.Name) ||
		database.Owner != DatabaseOwnerRole ||
		database.ExtensionOwnerApplicationLogin ||
		database.OwnershipAttestation != "bootstrap_live_catalog" ||
		database.Encoding != PostgresEncoding ||
		database.LocaleProvider != PostgresLocaleProvider ||
		database.Collation != PostgresCollation ||
		database.CType != PostgresCharacterType ||
		database.DataChecksums != PostgresDataChecksums ||
		database.PublicConnect || database.PublicCreate || database.PublicTemp {
		return fmt.Errorf("bootstrap authority database contract is invalid")
	}
	if origin == DatabaseOriginFresh {
		if !database.CreateDatabase || database.SourceOwner != nil ||
			database.ExtensionOwner != "harpy_extension_admin" ||
			database.ExtensionOwnerMode != "fresh_leased_admin" ||
			database.OwnershipExecutor != "cluster_bootstrap" ||
			database.Template == nil || *database.Template != "template0" {
			return fmt.Errorf("fresh database authority is invalid")
		}
		return nil
	}
	if database.CreateDatabase || database.SourceOwner == nil ||
		!validRoleName(*database.SourceOwner) ||
		database.OwnershipExecutor != "external_root_provisioner" ||
		database.Template != nil {
		return fmt.Errorf("adopted database authority is invalid")
	}
	switch database.ExtensionOwnerMode {
	case "adopted_reassigned":
		if database.ExtensionOwner != "harpy_legacy_extension_owner" {
			return fmt.Errorf("reassigned extension owner is invalid")
		}
	case "adopted_retained_inventory":
		if !validRoleName(database.ExtensionOwner) ||
			strings.HasPrefix(database.ExtensionOwner, "harpy_") {
			return fmt.Errorf("retained extension owner is invalid")
		}
	default:
		return fmt.Errorf("adopted extension owner mode is invalid")
	}
	return nil
}
