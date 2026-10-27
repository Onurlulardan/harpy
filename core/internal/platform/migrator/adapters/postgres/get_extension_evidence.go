package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

//go:embed sql/get_extension_evidence.sql
var getExtensionEvidenceSQL string

func GetExtensionEvidence(
	ctx context.Context,
	database goosedb.DBTxConn,
	name, version, verifiedRunID string,
) (ExtensionEvidenceRecord, error) {
	record := ExtensionEvidenceRecord{
		Name: name, Version: version, VerifiedRunID: verifiedRunID,
	}
	var capabilitiesJSON string
	if err := database.QueryRowContext(
		ctx, getExtensionEvidenceSQL, name, version, verifiedRunID,
	).Scan(
		&record.SchemaName, &record.OwnerRole, &record.ControlFileSHA256,
		&capabilitiesJSON, &record.DatabaseOrigin,
	); err != nil {
		return ExtensionEvidenceRecord{}, fmt.Errorf("get extension %s evidence: %w", name, err)
	}
	capabilities, err := decodeExtensionCapabilities(capabilitiesJSON)
	if err != nil {
		return ExtensionEvidenceRecord{}, fmt.Errorf("get extension %s evidence: %w", name, err)
	}
	record.Capabilities = capabilities
	return record, nil
}

func decodeExtensionCapabilities(document string) ([]string, error) {
	var capabilities []string
	if err := json.Unmarshal([]byte(document), &capabilities); err != nil {
		return nil, fmt.Errorf("decode extension capabilities: %w", err)
	}
	if capabilities == nil {
		return nil, fmt.Errorf("decode extension capabilities: null array")
	}
	return capabilities, nil
}
