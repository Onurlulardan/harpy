package packagecontract

import (
	"bytes"
	"encoding/json"
	"io"
)

const MaximumManifestBytes = 1 << 20

type manifestWire struct {
	Contract         string           `json:"contract"`
	PackageID        string           `json:"package_id"`
	Slug             string           `json:"slug"`
	Version          string           `json:"version"`
	PlatformContract string           `json:"platform_contract"`
	Profile          Profile          `json:"profile"`
	Dependencies     []dependencyWire `json:"dependencies"`
	Components       []componentWire  `json:"components"`
	Menu             []menuWire       `json:"menu"`
}

type dependencyWire struct {
	PackageID string         `json:"package_id"`
	Range     string         `json:"range"`
	Kind      DependencyKind `json:"kind"`
}

type componentWire struct {
	ID, Owner, Contract, ContentSHA256 string
	Kind                               ComponentKind
}

func (value *componentWire) UnmarshalJSON(data []byte) error {
	type wire struct {
		ID, Owner, Contract string
		Kind                ComponentKind `json:"kind"`
		ContentSHA256       string        `json:"content_sha256"`
	}
	var decoded wire
	if err := strictDecode(data, &decoded); err != nil {
		return err
	}
	value.ID, value.Owner, value.Contract = decoded.ID, decoded.Owner, decoded.Contract
	value.Kind, value.ContentSHA256 = decoded.Kind, decoded.ContentSHA256
	return nil
}

type menuWire struct {
	ID, ParentID, TargetID, Label, URL string
	Kind                               MenuKind `json:"kind"`
	Order                              int      `json:"order"`
}

func (value *menuWire) UnmarshalJSON(data []byte) error {
	type wire struct {
		ID, ParentID, TargetID, Label, URL string
		Kind                               MenuKind `json:"kind"`
		Order                              int      `json:"order"`
	}
	var decoded wire
	if err := strictDecode(data, &decoded); err != nil {
		return err
	}
	*value = menuWire(decoded)
	return nil
}

func Decode(document []byte) (Manifest, error) {
	if len(document) == 0 || len(document) > MaximumManifestBytes {
		return Manifest{}, ErrManifestInvalid
	}
	var wire manifestWire
	if err := strictDecode(document, &wire); err != nil {
		return Manifest{}, ErrManifestInvalid
	}
	return buildManifest(wire, document)
}

func strictDecode(document []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrManifestInvalid
	}
	return nil
}
