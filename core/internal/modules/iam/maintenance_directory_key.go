package iam

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
)

const maintenanceDirectoryKeyRedacted = "[REDACTED]"

type MaintenanceDirectoryKey struct {
	id       string
	material [32]byte
	active   bool
}

func NewMaintenanceDirectoryKey(
	id string,
	material [32]byte,
	active bool,
) (MaintenanceDirectoryKey, error) {
	defer clear(material[:])
	if !validMaintenanceDirectoryKeyID(id) || zeroDirectoryMaterial(material) {
		return MaintenanceDirectoryKey{}, fmt.Errorf(
			"IAM maintenance directory key: %w",
			ErrMaintenanceDirectoryKeyInvalid,
		)
	}
	return MaintenanceDirectoryKey{id: id, material: material, active: active}, nil
}

func validMaintenanceDirectoryKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}

func (key MaintenanceDirectoryKey) ID() string   { return key.id }
func (key MaintenanceDirectoryKey) Active() bool { return key.active }

func (key MaintenanceDirectoryKey) String() string {
	return fmt.Sprintf(
		"MaintenanceDirectoryKey{ID:%q Material:%s Active:%t}",
		key.id, maintenanceDirectoryKeyRedacted, key.active,
	)
}

func (key MaintenanceDirectoryKey) GoString() string { return key.String() }

func (key MaintenanceDirectoryKey) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, key.String())
}

func (key MaintenanceDirectoryKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string `json:"id"`
		Material string `json:"material"`
		Active   bool   `json:"active"`
	}{ID: key.id, Material: maintenanceDirectoryKeyRedacted, Active: key.active})
}

func (key MaintenanceDirectoryKey) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", key.id),
		slog.String("material", maintenanceDirectoryKeyRedacted),
		slog.Bool("active", key.active),
	)
}

func zeroDirectoryMaterial(material [32]byte) bool {
	var aggregate byte
	for _, value := range material {
		aggregate |= value
	}
	return aggregate == 0
}
