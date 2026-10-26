package devjournal

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type journalBinding struct {
	proof                 artifact.VerifiedBindingProof
	signerGeneration      int64
	signerKID             string
	signerPublicKeySHA256 string
}

type journalAck struct {
	ack artifact.VerifiedAck
}

type journalState struct {
	bindings     map[string]journalBinding
	environments map[string]string
	heads        map[string]ports.ExternalHead
	operations   map[string]journalAck
	sequence     int64
}

func (registry *Registry) loadState(ctx context.Context) (journalState, error) {
	records, err := readRecords(registry.path)
	state := journalState{
		bindings: map[string]journalBinding{}, environments: map[string]string{},
		heads: map[string]ports.ExternalHead{}, operations: map[string]journalAck{},
	}
	if err != nil {
		return state, err
	}
	public := registry.privateKey.Public().(ed25519.PublicKey)
	for _, record := range records {
		bytes, decodeErr := base64.RawURLEncoding.DecodeString(record.Artifact)
		if decodeErr != nil || base64.RawURLEncoding.EncodeToString(bytes) != record.Artifact {
			return state, fmt.Errorf("development authority journal encoding is invalid")
		}
		if record.Kind == "binding" {
			if err := registry.applyBindingRecord(ctx, &state, record, bytes); err != nil {
				return state, err
			}
		} else if record.Kind == "ack" {
			if err := registry.applyAckRecord(ctx, &state, record, bytes, public); err != nil {
				return state, err
			}
		} else {
			return state, fmt.Errorf("development authority journal kind is invalid")
		}
	}
	return state, nil
}

func headMapKey(bindingID string, kind artifact.ScopeKind, scopeID string) string {
	return bindingID + "\x00" + string(kind) + "\x00" + scopeID
}
