package recoverykeymaterial

import (
	"errors"
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
)

const (
	Schema               = "recovery-key-provider-material.v1"
	MaximumDocumentBytes = 32 * 1024

	PurposeComponentSealing Purpose = "recovery_component_sealing"
	ProviderLocalKeyring            = "local_keyring"
	ProviderForgeGuard              = "forgeguard"
)

var ErrInvalid = errors.New("recovery key material is invalid")

type Purpose string

type Expected struct {
	Purpose        Purpose
	ProviderKind   string
	ProviderHandle string
	EnvironmentID  string
	Generation     int64
}

type keyWire struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	Bytes     string `json:"key_bytes_base64url"`
}

type materialBody struct {
	DRKEK             keyWire `json:"dr_kek"`
	RecoveryPublicKey keyWire `json:"recovery_public_key"`
}

type materialWire struct {
	Schema         string       `json:"schema"`
	Purpose        Purpose      `json:"purpose"`
	ProviderKind   string       `json:"provider_kind"`
	ProviderHandle string       `json:"provider_handle"`
	EnvironmentID  string       `json:"environment_id"`
	Generation     int64        `json:"generation"`
	Synthetic      bool         `json:"synthetic"`
	Material       materialBody `json:"material"`
	MaterialSHA256 string       `json:"material_sha256"`
}

type Material struct {
	purpose        Purpose
	providerKind   string
	providerHandle string
	environmentID  string
	generation     int64
	drKeyID        string
	recoveryKeyID  string
	documentSHA256 string
	drKEK          [32]byte
	recoveryPublic [32]byte
}

func (value Material) Purpose() Purpose       { return value.purpose }
func (value Material) ProviderKind() string   { return value.providerKind }
func (value Material) ProviderHandle() string { return value.providerHandle }
func (value Material) EnvironmentID() string  { return value.environmentID }
func (value Material) Generation() int64      { return value.generation }
func (value Material) DRKeyID() string        { return value.drKeyID }
func (value Material) RecoveryKeyID() string  { return value.recoveryKeyID }
func (value Material) DocumentSHA256() string { return value.documentSHA256 }

func (value Material) SealKeys(random io.Reader) backupcrypto.SealKeys {
	return backupcrypto.SealKeys{
		DRKEK:          append([]byte(nil), value.drKEK[:]...),
		RecoveryPublic: append([]byte(nil), value.recoveryPublic[:]...),
		Random:         random,
	}
}

func (value Material) DROpenKeys() backupcrypto.OpenKeys {
	return backupcrypto.OpenKeys{
		DRKEK: append([]byte(nil), value.drKEK[:]...),
	}
}

func (value *Material) Clear() {
	if value == nil {
		return
	}
	clear(value.drKEK[:])
	clear(value.recoveryPublic[:])
}
