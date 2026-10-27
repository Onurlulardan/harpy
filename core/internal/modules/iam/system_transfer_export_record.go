package iam

import (
	"bytes"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

type systemTransferExportRecord struct {
	id      uuid.UUID
	members map[string]json.RawMessage
}

func decodeSystemTransferExportRecord(
	document []byte,
	fields []string,
) (systemTransferExportRecord, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
	}
	members := make(map[string]json.RawMessage, len(fields))
	for decoder.More() {
		token, tokenErr := decoder.Token()
		name, validName := token.(string)
		if tokenErr != nil || !validName {
			return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
		}
		if _, duplicate := members[name]; duplicate {
			return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
		}
		members[name] = append(json.RawMessage(nil), value...)
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(members) != len(fields) {
		return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
	}
	if _, err = decoder.Token(); err != io.EOF {
		return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
	}
	for _, field := range fields {
		if _, present := members[field]; !present {
			return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
		}
	}
	var encodedID string
	rawID := members["id"]
	identifier, parseErr := uuid.Nil, error(nil)
	if json.Unmarshal(rawID, &encodedID) == nil {
		identifier, parseErr = uuid.Parse(encodedID)
	}
	canonicalID, marshalErr := json.Marshal(encodedID)
	if parseErr != nil || marshalErr != nil || identifier == uuid.Nil ||
		identifier.Variant() != uuid.RFC4122 || identifier.String() != encodedID ||
		!bytes.Equal(canonicalID, rawID) {
		return systemTransferExportRecord{}, systemtransfer.ErrExportQueryInvalid
	}
	return systemTransferExportRecord{id: identifier, members: members}, nil
}

func systemTransferExportCell(
	kind queryapp.ExportLogicalType,
	raw json.RawMessage,
) (systemtransfer.ExportCell, error) {
	if bytes.Equal(raw, []byte("null")) {
		return systemtransfer.ExportCell{Kind: systemtransfer.ExportCellNull}, nil
	}
	if kind == queryapp.ExportLogicalTypeBoolean {
		if !bytes.Equal(raw, []byte("true")) && !bytes.Equal(raw, []byte("false")) {
			return systemtransfer.ExportCell{}, systemtransfer.ErrExportQueryInvalid
		}
		return systemtransfer.ExportCell{Kind: systemtransfer.ExportCellBoolean, Text: string(raw)}, nil
	}
	if kind == queryapp.ExportLogicalTypeJSON || kind == queryapp.ExportLogicalTypeTextList ||
		kind == queryapp.ExportLogicalTypeUUIDList {
		canonical, err := canonicaljsonv1.Canonicalize(raw, 1<<20)
		if err != nil {
			return systemtransfer.ExportCell{}, systemtransfer.ErrExportQueryInvalid
		}
		return systemtransfer.ExportCell{Kind: systemtransfer.ExportCellJSON, Text: string(canonical)}, nil
	}
	var scalar string
	if json.Unmarshal(raw, &scalar) != nil {
		return systemtransfer.ExportCell{}, systemtransfer.ErrExportQueryInvalid
	}
	canonical, err := json.Marshal(scalar)
	if err != nil || !bytes.Equal(canonical, raw) || !kind.Valid() {
		return systemtransfer.ExportCell{}, systemtransfer.ErrExportQueryInvalid
	}
	return systemtransfer.ExportCell{Kind: systemtransfer.ExportCellText, Text: scalar}, nil
}
