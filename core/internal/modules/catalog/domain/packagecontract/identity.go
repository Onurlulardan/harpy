package packagecontract

import (
	"encoding/hex"
	"fmt"
)

type StableID [16]byte
type PackageID struct{ value StableID }
type ComponentID struct{ value StableID }
type MenuItemID struct{ value StableID }

func ParseStableID(raw string) (StableID, error) {
	if len(raw) != 36 || raw[8] != '-' || raw[13] != '-' || raw[18] != '-' || raw[23] != '-' {
		return StableID{}, ErrManifestInvalid
	}
	compact := raw[0:8] + raw[9:13] + raw[14:18] + raw[19:23] + raw[24:36]
	decoded, err := hex.DecodeString(compact)
	if err != nil || hex.EncodeToString(decoded) != compact {
		return StableID{}, ErrManifestInvalid
	}
	var value StableID
	copy(value[:], decoded)
	if !value.Valid() {
		return StableID{}, ErrManifestInvalid
	}
	return value, nil
}

func ParsePackageID(raw string) (PackageID, error) {
	value, err := ParseStableID(raw)
	return PackageID{value: value}, err
}

func ParseComponentID(raw string) (ComponentID, error) {
	value, err := ParseStableID(raw)
	return ComponentID{value: value}, err
}

func ParseMenuItemID(raw string) (MenuItemID, error) {
	value, err := ParseStableID(raw)
	return MenuItemID{value: value}, err
}

func (value StableID) String() string {
	raw := hex.EncodeToString(value[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", raw[:8], raw[8:12], raw[12:16], raw[16:20], raw[20:])
}

func (value StableID) Valid() bool {
	var zero StableID
	return value != zero
}

func (value PackageID) String() string   { return value.value.String() }
func (value ComponentID) String() string { return value.value.String() }
func (value MenuItemID) String() string  { return value.value.String() }

func (value PackageID) Valid() bool   { return value.value.Valid() }
func (value ComponentID) Valid() bool { return value.value.Valid() }
func (value MenuItemID) Valid() bool  { return value.value.Valid() }
