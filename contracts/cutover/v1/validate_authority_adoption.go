package cutoverv1

import (
	"strings"
	"time"
)

const (
	authorityOwner     = "harpy_cutover_authority_owner"
	emptyCatalogSHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

var authorityObjectOrder = [][4]string{
	{"schema", "cutover_authority", "cutover_authority", ""},
	{"table", "cutover_authority", "authority_adoption_inventory", ""},
	{"table", "cutover_authority", "finalization_authorizations", ""},
	{"table", "cutover_authority", "finalization_receipts", ""},
	{"table", "cutover_authority", "full_rollout_completion_receipts", ""},
	{"function", "cutover_authority", "catalog_definition_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "catalog_acl_sha256_v1", "text,oid"},
	{"function", "cutover_authority", "assert_adopted_catalog_v1", "text"},
	{"function", "cutover_authority", "finalization_state_sha256_v1", "uuid,uuid"},
	{"function", "cutover_authority", "record_credential_receipts_v1",
		"uuid,uuid,text,bytea,text,bytea"},
	{"function", "cutover_authority", "observe_finalizer_authority_v1", "uuid,uuid"},
	{"function", "cutover_authority", "assert_completed_evidence_v1", "uuid,uuid,bytea"},
	{"function", "cutover_authority", "assert_tenant_completion_v1", "uuid,uuid"},
	{"function", "cutover_authority", "assert_terminal_evidence_v1",
		"uuid,uuid,text,bytea"},
	{"function", "cutover_authority", "validate_finalization_envelope_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "authorize_finalization_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "consume_finalization_authorization_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "finalize_operation_v1",
		"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea"},
	{"function", "cutover_authority", "verify_finalization_receipt_v1",
		"uuid,uuid,uuid,text,text"},
	{"function", "cutover_authority", "record_full_rollout_completion_v1",
		"uuid,uuid,uuid,text,text,text,text,text,text,text,text,text,timestampwithtimezone,timestampwithtimezone,timestampwithtimezone"},
}

func (value AuthorityAdoptionOperation) Validate() error {
	if value.ContractVersion != "cutover-authority-adoption-operation.v1" ||
		!validUUID(value.OperationID) || !validUUID(value.ReleaseID) ||
		!validUUID(value.EnvironmentID) || !validDigest(value.MigrationManifestSHA256) ||
		!validDigest(value.BootstrapAuthoritySHA256) ||
		!validDigest(value.PreCatalogContractSHA256) ||
		!validDigest(value.PostCatalogContractSHA256) {
		return contractError("authority adoption operation header")
	}
	issuedAt, issuedErr := parseUTC(value.IssuedAt)
	expiresAt, expiresErr := parseUTC(value.ExpiresAt)
	if issuedErr != nil || expiresErr != nil || !issuedAt.Before(expiresAt) ||
		expiresAt.Sub(issuedAt) > 15*time.Minute || ValidateDatabaseIdentity(value.Target) != nil {
		return contractError("authority adoption operation validity")
	}
	if value.Executor != "external_root_provisioner" ||
		!sqlNamePattern.MatchString(value.ExecutorSessionUser) ||
		strings.HasPrefix(value.ExecutorSessionUser, "harpy_") ||
		value.MembershipPolicy != "always_zero" ||
		!validAuthorityOwnershipMode(value.Mode, value.OwnershipMode) ||
		validateAuthorityOwnerRole(value.AuthorityOwnerRole, len(value.Objects)) != nil {
		return contractError("authority adoption executor")
	}
	if value.PostCatalogContractSHA256 != AuthorityOperationCatalogContractSHA256(value.Objects) {
		return contractError("authority adoption catalog contract digest")
	}
	if err := validateAuthorityDependencies(
		value.Dependencies, value.DependencyContractSHA256,
	); err != nil {
		return err
	}
	if err := validateAuthorityQuiescencePolicy(value.QuiescencePolicy, value.Objects); err != nil {
		return err
	}
	return validateAuthorityMode(value)
}

func validateAuthorityMode(value AuthorityAdoptionOperation) error {
	switch value.Mode {
	case "initial_adoption":
		if value.PreviousCatalogGeneration != 0 || value.CatalogGeneration != 1 ||
			value.PreCatalogContractSHA256 != emptyCatalogSHA256 ||
			value.PreviousAuthorityReceiptSHA256 != nil || value.UpgradeArtifactSHA256 != nil ||
			len(value.Objects) != len(authorityObjectOrder) {
			return contractError("initial authority adoption")
		}
		return validateAuthorityObjects(value.Objects, "adopt", value.CatalogGeneration)
	case "reaffirmation":
		if value.CatalogGeneration < 1 || value.PreviousCatalogGeneration != value.CatalogGeneration ||
			value.PreCatalogContractSHA256 != value.PostCatalogContractSHA256 ||
			!validOptionalDigest(value.PreviousAuthorityReceiptSHA256) ||
			value.UpgradeArtifactSHA256 != nil {
			return contractError("authority reaffirmation")
		}
		return validateAuthorityObjects(value.Objects, "retain", value.CatalogGeneration)
	case "catalog_upgrade":
		if value.PreviousCatalogGeneration < 1 ||
			value.CatalogGeneration != value.PreviousCatalogGeneration+1 ||
			value.PreCatalogContractSHA256 == value.PostCatalogContractSHA256 ||
			!validOptionalDigest(value.PreviousAuthorityReceiptSHA256) ||
			!validOptionalDigest(value.UpgradeArtifactSHA256) ||
			len(value.Objects) <= len(authorityObjectOrder) {
			return contractError("authority catalog upgrade")
		}
		return validateUpgradeObjects(value.Objects, value.CatalogGeneration)
	default:
		return contractError("authority adoption mode")
	}
}

func validateAuthorityObjects(objects []AuthorityAdoptionObject, action string, generation int) error {
	if len(objects) < len(authorityObjectOrder) {
		return contractError("authority adoption object set")
	}
	seen := make(map[string]struct{}, len(objects))
	previousGeneration := 0
	for index, object := range objects {
		if !validAuthorityObjectIdentity(object, index, seen) ||
			object.IntroducedGeneration < previousGeneration {
			return contractError("authority adoption object identity")
		}
		previousGeneration = object.IntroducedGeneration
		preOwner := authorityOwner
		if action == "adopt" {
			preOwner = "harpy_owner"
		}
		if object.Action != action || object.PreAdoptionOwner != preOwner ||
			object.AdoptedOwner != authorityOwner || object.IntroducedGeneration < 1 ||
			object.IntroducedGeneration > generation || !validAuthorityObjectHashes(object) ||
			(action == "retain" && object.PreAdoptionACLSHA256 != object.AdoptedACLSHA256) {
			return contractError("authority adoption object")
		}
	}
	return nil
}

func validateUpgradeObjects(objects []AuthorityAdoptionObject, generation int) error {
	seen := make(map[string]struct{}, len(objects))
	previousGeneration, added := 0, false
	for index, object := range objects {
		if !validAuthorityObjectIdentity(object, index, seen) ||
			object.IntroducedGeneration < previousGeneration {
			return contractError("authority upgrade object identity")
		}
		previousGeneration = object.IntroducedGeneration
		if object.IntroducedGeneration < generation {
			if added || object.Action != "retain" || object.PreAdoptionOwner != authorityOwner ||
				object.AdoptedOwner != authorityOwner || !validAuthorityObjectHashes(object) ||
				object.PreAdoptionACLSHA256 != object.AdoptedACLSHA256 {
				return contractError("authority upgrade retained object")
			}
			continue
		}
		added = true
		if object.IntroducedGeneration != generation || object.Action != "add" ||
			object.PreAdoptionOwner != authorityOwner || object.AdoptedOwner != authorityOwner ||
			!validAuthorityObjectHashes(object) {
			return contractError("authority upgrade added object")
		}
	}
	if !added {
		return contractError("authority upgrade has no additions")
	}
	return nil
}

func validAuthorityOwnershipMode(mode, ownershipMode string) bool {
	if mode == "catalog_upgrade" {
		return ownershipMode == "superuser_set_local_role_append_only"
	}
	return ownershipMode == "superuser_direct_per_object"
}

func validAuthorityObjectHashes(object AuthorityAdoptionObject) bool {
	return validDigest(object.PreDefinitionSHA256) &&
		validDigest(object.AdoptedDefinitionSHA256) &&
		validDigest(object.PreAdoptionACLSHA256) && validDigest(object.AdoptedACLSHA256) &&
		object.PreDefinitionSHA256 == object.AdoptedDefinitionSHA256
}

func matchesBaseAuthorityObject(object AuthorityAdoptionObject, index int) bool {
	expected := authorityObjectOrder[index]
	return object.Kind == expected[0] && object.SchemaName == expected[1] &&
		object.ObjectName == expected[2] && object.IdentityArguments == expected[3]
}

func validOptionalDigest(value *string) bool { return value != nil && validDigest(*value) }

func validateAuthorityOwnerRole(role AuthorityOwnerRoleState, objectCount int) error {
	if role.Login || role.Inherit || role.Superuser || role.CreateDB || role.CreateRole ||
		role.Replication || role.BypassRLS || role.ConnectionLimit != -1 ||
		role.PasswordState != "null" || role.ValidUntil != "infinity" ||
		role.MembershipCount != 0 || role.OwnedObjectCount != objectCount {
		return contractError("authority owner role state")
	}
	return nil
}
