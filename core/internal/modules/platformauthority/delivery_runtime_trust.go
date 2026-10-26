package platformauthority

import (
	"crypto/ed25519"
	"fmt"
	"io"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const maximumDeliveryArtifactBytes = 64 * 1024

func verifyDeliveryTrust(
	dependencies DeliveryDependencies,
	kind ports.RegistryKind,
) (ports.VerifiedTrustBundle, error) {
	ackDocument, err := readDeliveryArtifact(dependencies.AckTrustPath)
	if err != nil {
		return ports.VerifiedTrustBundle{}, err
	}
	defer clear(ackDocument)
	rootDocument, err := readDeliveryArtifact(dependencies.ReleaseTrustPath)
	if err != nil {
		return ports.VerifiedTrustBundle{}, err
	}
	defer clear(rootDocument)
	root, err := attestation.ParseTrustRoot(rootDocument)
	if err != nil {
		return ports.VerifiedTrustBundle{}, fmt.Errorf(
			"parse authority release trust root: %w", err,
		)
	}
	var selected ed25519.PublicKey
	for _, candidate := range root.ActivePublicKeys(dependencies.Clock.Now()) {
		if _, verifyErr := ports.VerifyTrustBundle(
			ackDocument, kind, dependencies.Clock.Now(), candidate,
		); verifyErr == nil {
			if selected != nil {
				return ports.VerifiedTrustBundle{},
					fmt.Errorf("authority release trust root is ambiguous")
			}
			selected = append(ed25519.PublicKey(nil), candidate...)
		}
	}
	if selected == nil {
		return ports.VerifiedTrustBundle{},
			fmt.Errorf("authority release trust root has no matching key")
	}
	return ports.VerifyTrustBundle(
		ackDocument, kind, dependencies.Clock.Now(), selected,
	)
}

func readDeliveryArtifact(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open authority artifact: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 2 ||
		info.Size() > maximumDeliveryArtifactBytes {
		return nil, fmt.Errorf("authority artifact is invalid")
	}
	document := make([]byte, info.Size())
	if _, err = io.ReadFull(file, document); err != nil {
		clear(document)
		return nil, fmt.Errorf("read authority artifact: %w", err)
	}
	return document, nil
}
