package runtimev1

import "fmt"

func ValidateExtensionReceipt(receipt ExtensionReceipt) error {
	payload := receipt.Payload
	if receipt.ContractVersion != "postgres-extension-receipt.v1" ||
		receipt.Kind != "extension_admin" {
		return fmt.Errorf("extension receipt identity is invalid")
	}
	for _, identity := range []struct {
		value string
		field string
	}{
		{payload.RunID, "run_id"}, {payload.ReleaseID, "release_id"},
		{payload.CredentialLeaseID, "credential_lease_id"},
	} {
		if err := validateUUID(identity.value, identity.field); err != nil {
			return err
		}
	}
	if _, err := validateTimestamp(payload.IssuedAt, "issued_at"); err != nil {
		return err
	}
	if err := validateReceiptDigests(
		receiptDigest{payload.ReleaseManifestSHA256, "release_manifest_sha256"},
		receiptDigest{payload.BootstrapAuthorityManifestSHA256, "bootstrap_authority_manifest_sha256"},
		receiptDigest{payload.OwnershipManifestSHA256, "ownership_manifest_sha256"},
		receiptDigest{payload.OperationDigest, "operation_digest"},
		receiptDigest{payload.ExtensionManifestSHA256, "extension_manifest_sha256"},
	); err != nil {
		return err
	}
	origin, err := validateReceiptDatabase(
		payload.DatabaseName, payload.DatabaseSystemIdentifier, payload.DatabaseOrigin,
		payload.DatabaseOwner, payload.DatabaseEncoding, payload.DatabaseLocaleProvider,
		payload.DatabaseCollation, payload.DatabaseCType, payload.DatabaseOID,
		payload.DataChecksums,
	)
	if err != nil {
		return err
	}
	if payload.PostgresImageDigest != PostgresImageDigest ||
		payload.ServerVersionNumber != PostgresServerVersionNumber ||
		payload.SessionUser != "harpy_extension_admin" ||
		payload.CurrentUser != "harpy_extension_admin" || payload.Result != "succeeded" {
		return fmt.Errorf("extension receipt execution contract is invalid")
	}
	if err := validateExtensionRecords(payload.Extensions, origin); err != nil {
		return err
	}
	return ValidateAttestationSignature(receipt.Signature)
}

func validateExtensionRecords(records []ExtensionRecord, origin DatabaseOrigin) error {
	if len(records) == 0 {
		return fmt.Errorf("extension receipt must contain at least one extension")
	}
	previous := ""
	for index, record := range records {
		key := record.Name + "\x00" + record.Version + "\x00" + record.Schema
		if !databaseNamePattern.MatchString(record.Name) ||
			!extensionVersion.MatchString(record.Version) ||
			!databaseNamePattern.MatchString(record.Schema) ||
			!validExtensionOwner(record.Owner) {
			return fmt.Errorf("extension record %d identity is invalid", index)
		}
		if origin == DatabaseOriginFresh &&
			(record.Schema != "platform_ext" || record.Owner != "harpy_extension_admin") {
			return fmt.Errorf("fresh extension record %d authority is invalid", index)
		}
		if previous != "" && key <= previous {
			return fmt.Errorf("extension records must be unique and sorted")
		}
		previous = key
		if err := validateDigest(record.ControlFileSHA256, "control_file_sha256"); err != nil {
			return err
		}
	}
	return nil
}
