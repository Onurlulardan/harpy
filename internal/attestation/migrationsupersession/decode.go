package migrationsupersession

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func DecodeCatalog(document []byte) (Catalog, error) {
	if len(document) == 0 || len(document) > maximumCatalogBytes {
		return Catalog{}, fmt.Errorf("migration supersession catalog size is invalid")
	}
	if document[len(document)-1] != '\n' {
		return Catalog{}, fmt.Errorf("migration supersession catalog terminator is invalid")
	}
	payload := document[:len(document)-1]
	canonical, err := attestation.Canonicalize(payload)
	if err != nil {
		return Catalog{}, fmt.Errorf("canonicalize migration supersession catalog: %w", err)
	}
	if !bytes.Equal(payload, canonical) {
		return Catalog{}, fmt.Errorf("migration supersession catalog is not canonical JSON")
	}
	var decoded catalogDocument
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return Catalog{}, fmt.Errorf("decode migration supersession catalog: %w", err)
	}
	catalog := Catalog{
		contractVersion: decoded.ContractVersion,
		owner:           decoded.Owner,
		scope:           decoded.Scope,
		sourceDirectory: decoded.SourceDirectory,
		entries:         make([]Entry, len(decoded.Entries)),
	}
	for index, entry := range decoded.Entries {
		catalog.entries[index] = Entry{
			oldPath: entry.OldPath, oldSHA256: entry.OldSHA256,
			replacementPath:   entry.ReplacementPath,
			replacementSHA256: entry.ReplacementSHA256,
		}
	}
	if err := validateCatalog(catalog); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}
