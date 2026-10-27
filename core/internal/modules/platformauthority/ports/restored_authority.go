package ports

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type RestoredAuthorityEvidence struct {
	EnvironmentID             string
	ProtectedHeadSHA256       string
	ProtectedAckSHA256        string
	ProtectedCheckpointSHA256 string
	AckTrustSHA256            string
	ProtectedAck              []byte
	ProtectedCheckpoint       []byte
}

func VerifyRestoredAuthority(
	evidence RestoredAuthorityEvidence,
	trustDocument []byte,
	now time.Time,
) error {
	if digestHex(evidence.ProtectedAck) != evidence.ProtectedAckSHA256 ||
		digestHex(evidence.ProtectedCheckpoint) != evidence.ProtectedCheckpointSHA256 {
		return fmt.Errorf("protected authority digest is invalid")
	}
	ack, err := artifact.ParseAck(evidence.ProtectedAck)
	if err != nil || ack.Value.ScopeKind != artifact.ScopeEnvironment ||
		ack.Value.ScopeID != evidence.EnvironmentID ||
		ack.Value.StateSHA256 != evidence.ProtectedHeadSHA256 {
		return fmt.Errorf("protected authority ack binding is invalid")
	}
	checkpoint, err := artifact.ParseCheckpoint(evidence.ProtectedCheckpoint)
	if err != nil || checkpoint.Value.RegistryKind != ack.Value.RegistryKind ||
		checkpoint.Value.RegistrySequence < ack.Value.RegistrySequence {
		return fmt.Errorf("protected authority checkpoint is stale")
	}
	trust, err := artifact.VerifyPinnedTrustBundle(
		trustDocument, ack.Value.RegistryKind, evidence.AckTrustSHA256, now,
	)
	if err != nil {
		return fmt.Errorf("protected authority trust is invalid")
	}
	if _, err = artifact.VerifyAckPinned(evidence.ProtectedAck, trust, now); err != nil {
		return err
	}
	_, err = artifact.VerifyCheckpoint(evidence.ProtectedCheckpoint, trust, now)
	return err
}

func digestHex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
