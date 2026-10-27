package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func statementSeal(statement RecordStatement) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.record-statement.v1")
	writer.writeString(statement.sql)
	writer.writeInt(int64(statement.requestedSize))
	writer.writeInt(int64(statement.class))
	writer.writeInt(int64(statement.countMode))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeStatementColumns(writer, statement.columns) != nil ||
		writeStatementBinding(writer, statement.binding) != nil ||
		writeCursorContext(writer, statement.cursorContext) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}

func writeStatementArguments(writer *sealWriter, arguments []any) error {
	writer.writeInt(int64(len(arguments)))
	for _, argument := range arguments {
		switch value := argument.(type) {
		case string:
			writer.writeString("string")
			writer.writeString(value)
		case bool:
			writer.writeString("bool")
			writer.writeBool(value)
		case int:
			writer.writeString("int")
			writer.writeInt(int64(value))
		case int64:
			writer.writeString("int64")
			writer.writeInt(value)
		case []string:
			writer.writeString("strings")
			writer.writeInt(int64(len(value)))
			for _, item := range value {
				writer.writeString(item)
			}
		case []uuid.UUID:
			writer.writeString("uuids")
			writer.writeInt(int64(len(value)))
			for _, identifier := range value {
				if identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
					return ErrContract
				}
				writer.writeBytes(identifier[:])
			}
		default:
			return ErrContract
		}
	}
	return nil
}

func writeStatementColumns(writer *sealWriter, columns []ResultColumn) error {
	writer.writeInt(int64(len(columns)))
	for _, column := range columns {
		if column.Validate() != nil || writeFieldType(writer, column.fieldType) != nil {
			return ErrContract
		}
		writer.writeBytes(column.fieldID[:])
		writer.writeString(column.outputKey)
		writer.writeBool(column.hasOutput)
		writer.writeInt(int64(column.sortOrdinal))
		writer.writeInt(int64(column.sortDirection))
		writer.writeInt(int64(column.cursorCanonicalBytes))
	}
	return nil
}

func writeStatementBinding(writer *sealWriter, binding statementBinding) error {
	if binding.validate() != nil {
		return ErrContract
	}
	writer.writeInt(int64(binding.scope))
	writer.writeBytes(binding.scopeID[:])
	writer.writeBytes(binding.subject[:])
	writer.writeInt(int64(binding.audience))
	writer.writeString(binding.operationID)
	writer.writeInt(binding.snapshotUnix)
	writer.writeBytes(binding.authorization[:])
	writer.writeBytes(binding.mandatory[:])
	writer.writeBytes(binding.identity[:])
	writer.writeBytes(binding.continuation[:])
	return writeCatalogReference(writer, binding.catalog)
}

func writeCatalogReference(writer *sealWriter, reference api.CatalogReference) error {
	owner, ownerErr := reference.Owner()
	resource, resourceErr := reference.ResourceID()
	version, versionErr := reference.PublishedVersionID()
	catalogHash, hashErr := reference.CatalogHash()
	digest, digestErr := catalogHash.Bytes()
	if ownerErr != nil || resourceErr != nil || versionErr != nil ||
		hashErr != nil || digestErr != nil {
		return ErrContract
	}
	writer.writeString(owner)
	writer.writeBytes(resource[:])
	writer.writeBytes(version[:])
	writer.writeBytes(digest)
	return nil
}
