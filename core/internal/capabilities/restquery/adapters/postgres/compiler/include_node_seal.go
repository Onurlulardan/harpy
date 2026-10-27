package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func includeNodeSeal(node IncludeNode) ([sha256.Size]byte, error) {
	writer := newSealWriter("forgecrud.restquery.postgres.include-node.v1")
	path, pathErr := node.path.Value()
	targetStorage, targetStorageErr := node.targetStorage.FingerprintSHA256()
	relationStorage, relationStorageErr := node.relationStorage.FingerprintSHA256()
	if pathErr != nil || targetStorageErr != nil || relationStorageErr != nil ||
		writeIncludePlanBinding(writer, node.binding) != nil ||
		writeCatalogReference(writer, node.sourceCatalog) != nil ||
		writeCatalogReference(writer, node.targetCatalog) != nil ||
		writeIncludeStableField(writer, node.sourceStable) != nil ||
		writeIncludeStableField(writer, node.targetStable) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	writer.writeBytes(targetStorage[:])
	writer.writeBytes(relationStorage[:])
	writer.writeString(path)
	writer.writeString(node.relationName)
	writer.writeInt(int64(node.cardinality))
	writer.writeInt(int64(node.limit))
	writer.writeInt(int64(node.depth))
	writer.writeInt(int64(node.ordinal))
	writer.writeInt(int64(node.parentOrdinal))
	writer.writeInt(int64(node.siblingOrdinal))
	return writer.sum()
}

func writeIncludeStableField(writer *sealWriter, field interface {
	ID() (uuid.UUID, error)
	Type() (api.FieldType, error)
	Storage() (api.FieldStorage, error)
}) error {
	identifier, identifierErr := field.ID()
	fieldType, typeErr := field.Type()
	storage, storageErr := field.Storage()
	column, columnErr := storage.Column()
	fingerprint, fingerprintErr := sqlidentifier.Fingerprint(column)
	if identifierErr != nil || typeErr != nil || storageErr != nil ||
		columnErr != nil || fingerprintErr != nil ||
		writeFieldType(writer, fieldType) != nil {
		return ErrContract
	}
	writer.writeBytes(identifier[:])
	writer.writeBytes(fingerprint[:])
	return nil
}
