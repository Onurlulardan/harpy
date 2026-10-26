package runtimev1

import "fmt"

type receiptDigest struct {
	value string
	field string
}

func validateReceiptDatabase(
	name, systemIdentifier, originValue, owner, encoding,
	localeProvider, collation, ctype string,
	oid uint32,
	checksums bool,
) (DatabaseOrigin, error) {
	origin := DatabaseOrigin(originValue)
	if !databaseNamePattern.MatchString(name) || !origin.Valid() ||
		oid == 0 || owner != DatabaseOwnerRole || encoding != PostgresEncoding ||
		localeProvider != PostgresLocaleProvider || collation != PostgresCollation ||
		ctype != PostgresCharacterType || checksums != PostgresDataChecksums {
		return "", fmt.Errorf("receipt database identity is invalid")
	}
	if err := validateSystemIdentifier(systemIdentifier); err != nil {
		return "", err
	}
	return origin, nil
}

func validateReceiptDigests(digests ...receiptDigest) error {
	for _, digest := range digests {
		if err := validateDigest(digest.value, digest.field); err != nil {
			return err
		}
	}
	return nil
}

func validExtensionOwner(owner string) bool {
	if owner == "harpy_extension_admin" || owner == "harpy_legacy_extension_owner" {
		return true
	}
	return validRoleName(owner) && len(owner) > 10 && owner[:10] != "harpy_"
}
