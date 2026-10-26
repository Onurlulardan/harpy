package planner

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
)

const MaximumExplainBytes = 64 * 1024

var ErrExplainContract = errors.New("REST Query planner EXPLAIN contract is invalid")

// ParseExplainCost reads only the root plan's Total Cost from PostgreSQL JSON.
func ParseExplainCost(document []byte) (Cost, error) {
	if len(document) == 0 || len(document) > MaximumExplainBytes {
		return Cost{}, ErrExplainContract
	}
	scan, err := jsonscan.Scan(document, jsonscan.MaximumNesting)
	if err != nil || len(scan.DuplicatePointers()) != 0 {
		return Cost{}, ErrExplainContract
	}
	root, err := decodeExplainRoot(document)
	if err != nil || len(root) != 1 {
		return Cost{}, ErrExplainContract
	}
	wrapper, err := decodeRawObject(root[0])
	if err != nil || len(wrapper) != 1 {
		return Cost{}, ErrExplainContract
	}
	planDocument, exists := wrapper["Plan"]
	if !exists {
		return Cost{}, ErrExplainContract
	}
	plan, err := decodeRawObject(planDocument)
	if err != nil {
		return Cost{}, ErrExplainContract
	}
	encodedCost, exists := plan["Total Cost"]
	if !exists {
		return Cost{}, ErrExplainContract
	}
	return decodeCost(encodedCost)
}

func decodeExplainRoot(document []byte) ([]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var root []json.RawMessage
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	if err := requireJSONEnd(decoder); err != nil {
		return nil, err
	}
	return root, nil
}

func decodeRawObject(document []byte) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(document, &object); err != nil || object == nil {
		return nil, ErrExplainContract
	}
	return object, nil
}

func requireJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ErrExplainContract
	}
	return nil
}
