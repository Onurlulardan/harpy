package tabularmapping

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

var targetPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type mappingDocument struct {
	Contract string         `json:"contract"`
	Fields   []mappingField `json:"fields"`
}

type mappingField struct {
	Target     string          `json:"target"`
	Required   bool            `json:"required"`
	Expression json.RawMessage `json:"expression"`
}

type Program struct {
	document []byte
	contract string
	fields   []compiledField
}

func Compile(document []byte) (Program, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumMappingBytes)
	if err != nil {
		return Program{}, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	var wire mappingDocument
	if decoder.Decode(&wire) != nil {
		return Program{}, ErrInvalid
	}
	if _, err := decoder.Token(); err != io.EOF || !supportedContract(wire.Contract) ||
		len(wire.Fields) < 1 || len(wire.Fields) > MaximumFields {
		return Program{}, ErrInvalid
	}
	fields, err := compileFields(wire.Fields)
	if err != nil {
		return Program{}, err
	}
	return Program{document: canonical, contract: wire.Contract, fields: fields}, nil
}

func supportedContract(value string) bool {
	return value == ContractV2 || value == FormsImportContractV2
}

func compileFields(wire []mappingField) ([]compiledField, error) {
	fields := make([]compiledField, len(wire))
	targets := make(map[string]struct{}, len(wire))
	totalNodes := 0
	for index, field := range wire {
		if !targetPattern.MatchString(field.Target) || !utf8.ValidString(field.Target) {
			return nil, ErrInvalid
		}
		if _, duplicate := targets[field.Target]; duplicate {
			return nil, ErrInvalid
		}
		targets[field.Target] = struct{}{}
		compiled, nodes, err := compileExpression(field)
		if err != nil || totalNodes+nodes > MaximumNumericNodes {
			return nil, ErrInvalid
		}
		totalNodes += nodes
		fields[index] = compiled
	}
	return fields, nil
}

func (program Program) Contract() string { return program.contract }

func (program Program) Document() []byte { return append([]byte(nil), program.document...) }

func (program Program) Targets() []string {
	targets := make([]string, len(program.fields))
	for index, field := range program.fields {
		targets[index] = field.target
	}
	return targets
}
