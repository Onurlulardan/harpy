package formscleanup

import (
	"strings"

	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) scan(value row) (formscleanupapp.Snapshot, error) {
	snapshot, err := scanSnapshot(value)
	if err != nil && storage.notFound(err) {
		return formscleanupapp.Snapshot{}, formscleanupapp.ErrConflict
	}
	return snapshot, err
}

func scanSnapshot(value row, trailing ...any) (formscleanupapp.Snapshot, error) {
	var operationID, definitionID, state string
	var linkDigest, zeroDigest, receiptDigest []byte
	var completed dbInstant
	result := formscleanupapp.Snapshot{}
	destinations := []any{
		&operationID, &definitionID, &result.Epoch, &state,
		&result.LinkCount, &result.ItemCount, &result.VersionCount,
		&result.CiphertextBytes, &linkDigest, &zeroDigest, &receiptDigest,
		&result.RequestedAt, &completed,
	}
	if err := value.Scan(append(destinations, trailing...)...); err != nil {
		return formscleanupapp.Snapshot{}, err
	}
	parsedOperation, operationErr := domain.ParseID(operationID)
	parsedDefinition, definitionErr := domain.ParseID(definitionID)
	result.OperationID, result.DefinitionID = parsedOperation, parsedDefinition
	result.State = cleanupState(state)
	if completed.Valid {
		result.CompletedAt = completed.Time.UTC()
	}
	if operationErr != nil || definitionErr != nil || !result.OperationID.UUIDv7() ||
		!result.DefinitionID.UUIDv7() || result.Epoch < 2 ||
		result.LinkCount < 0 || result.ItemCount < 0 || result.VersionCount < 0 ||
		result.CiphertextBytes < 0 || result.State == "" ||
		!copyDigest(&result.LinkCatalogSHA256, linkDigest) ||
		!copyDigest(&result.ZeroBlockerSHA256, zeroDigest) ||
		!copyDigest(&result.ReceiptSHA256, receiptDigest) {
		return formscleanupapp.Snapshot{}, formscleanupapp.ErrUnavailable
	}
	return result, nil
}

func cleanupState(value string) formscleanupapi.State {
	switch strings.ToUpper(value) {
	case "REQUESTED":
		return formscleanupapi.StateRequested
	case "RUNNING":
		return formscleanupapi.StateRunning
	case "SUCCEEDED":
		return formscleanupapi.StateSucceeded
	default:
		return ""
	}
}

func copyDigest(target *[32]byte, source []byte) bool {
	if len(source) == 0 {
		return true
	}
	if len(source) != len(target) {
		return false
	}
	copy(target[:], source)
	return true
}
