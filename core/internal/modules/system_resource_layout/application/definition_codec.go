package application

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const maximumDefinitionBytes = 1 << 20

type definitionContractWire struct {
	Contract layoutapi.DefinitionContract `json:"contract"`
}

type definitionV2Wire struct {
	Contract layoutapi.DefinitionContract  `json:"contract"`
	Surfaces []layoutapi.SurfaceDefinition `json:"surfaces"`
}

func encodeDefinition(
	input layoutapi.Definition,
) (layoutapi.Definition, []byte, [sha256.Size]byte, error) {
	input.SourceContract = layoutapi.DefinitionContractV2
	definition := normalizeDefinition(input)
	if err := validateDefinition(definition); err != nil {
		return layoutapi.Definition{}, nil, [sha256.Size]byte{}, err
	}
	document, err := json.Marshal(definitionV2Wire{
		Contract: layoutapi.DefinitionContractV2, Surfaces: definition.Surfaces,
	})
	if err != nil {
		return layoutapi.Definition{}, nil, [sha256.Size]byte{}, layoutapi.ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, maximumDefinitionBytes)
	if err != nil {
		return layoutapi.Definition{}, nil, [sha256.Size]byte{}, layoutapi.ErrInvalid
	}
	return definition, canonical, sha256.Sum256(canonical), nil
}

func decodeDefinition(document []byte) (layoutapi.Definition, []byte, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, maximumDefinitionBytes)
	if err != nil {
		return layoutapi.Definition{}, nil, layoutapi.ErrInvalid
	}
	var contract definitionContractWire
	if json.Unmarshal(canonical, &contract) != nil {
		return layoutapi.Definition{}, nil, layoutapi.ErrInvalid
	}
	var definition layoutapi.Definition
	switch contract.Contract {
	case layoutapi.DefinitionContractV1:
		definition, err = decodeLegacyDefinition(canonical)
	case layoutapi.DefinitionContractV2:
		definition, err = decodeV2Definition(canonical)
	default:
		return layoutapi.Definition{}, nil, layoutapi.ErrInvalid
	}
	if err != nil || validateDefinition(definition) != nil {
		return layoutapi.Definition{}, nil, layoutapi.ErrInvalid
	}
	return definition, canonical, nil
}

func decodeV2Definition(document []byte) (layoutapi.Definition, error) {
	var wire definitionV2Wire
	if decodeExact(document, &wire) != nil ||
		wire.Contract != layoutapi.DefinitionContractV2 {
		return layoutapi.Definition{}, layoutapi.ErrInvalid
	}
	return normalizeDefinition(layoutapi.Definition{
		SourceContract: layoutapi.DefinitionContractV2, Surfaces: wire.Surfaces,
	}), nil
}

func decodeExact(document []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return layoutapi.ErrInvalid
	}
	return nil
}
