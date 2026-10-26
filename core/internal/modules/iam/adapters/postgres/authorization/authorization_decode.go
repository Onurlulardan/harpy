package postgres

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
)

func decodeAuthorizationSnapshot(
	actorJSON, resourceJSON, actionJSON []byte,
	clock int64,
	bindingsJSON, implicationsJSON []byte,
) (authorization.Snapshot, error) {
	actorDocument, err := decodeOptionalDocument[actorDocument](actorJSON)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	resourceDocument, err := decodeOptionalDocument[catalogDocument](resourceJSON)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	actionDocument, err := decodeOptionalDocument[catalogDocument](actionJSON)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	bindings, err := decodeDocumentArray[bindingDocument](bindingsJSON)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	implications, err := decodeDocumentArray[implicationDocument](implicationsJSON)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	snapshot, err := authorization.NewSnapshot(
		mapActorFact(actorDocument), mapResourceFact(resourceDocument),
		mapActionFact(actionDocument), clock,
		mapBindingFacts(bindings), mapImplicationFacts(implications),
	)
	if err != nil {
		return authorization.Snapshot{}, evidenceReadError(err)
	}
	return snapshot, nil
}

func decodeOptionalDocument[T any](document []byte) (*T, error) {
	if len(document) == 0 || bytes.Equal(document, []byte("null")) {
		return nil, nil
	}
	var result T
	if err := decodeExactJSON(document, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func decodeDocumentArray[T any](document []byte) ([]T, error) {
	if len(document) == 0 {
		return nil, errAuthorizationRead
	}
	var result []T
	if err := decodeExactJSON(document, &result); err != nil || result == nil {
		return nil, errAuthorizationRead
	}
	return result, nil
}

func decodeExactJSON(document []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errAuthorizationRead
	}
	return nil
}

func evidenceReadError(err error) error {
	if err == nil {
		return errAuthorizationRead
	}
	return fmt.Errorf("%w: %w", errAuthorizationRead, err)
}
