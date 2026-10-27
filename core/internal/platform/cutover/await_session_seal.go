package cutover

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func AwaitSessionSeal(
	ctx context.Context,
	path string,
	evidence AdoptionEvidence,
	now func() time.Time,
) (attestation.VerifiedCutoverSessionSeal, error) {
	if now == nil {
		return attestation.VerifiedCutoverSessionSeal{}, errors.New("cutover seal clock is required")
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return LoadSessionSeal(path, evidence, now().UTC())
		} else if !os.IsNotExist(err) {
			return attestation.VerifiedCutoverSessionSeal{}, errors.New("inspect cutover session seal")
		}
		if !now().Before(evidence.Lease.ExpiresAt()) {
			return attestation.VerifiedCutoverSessionSeal{}, errors.New("cutover lease expired before session seal")
		}
		select {
		case <-ctx.Done():
			return attestation.VerifiedCutoverSessionSeal{}, errors.New("cutover session seal wait cancelled")
		case <-ticker.C:
		}
	}
}
