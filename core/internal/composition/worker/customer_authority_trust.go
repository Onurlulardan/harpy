package worker

import (
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"time"

	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
)

const maximumCustomerAuthorityArtifactBytes = 64 * 1024

func readCustomerAuthorityArtifact(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open customer authority artifact: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 2 ||
		info.Size() > maximumCustomerAuthorityArtifactBytes {
		return nil, fmt.Errorf("customer authority artifact is invalid")
	}
	document := make([]byte, info.Size())
	if _, err := io.ReadFull(file, document); err != nil {
		clear(document)
		return nil, fmt.Errorf("read customer authority artifact: %w", err)
	}
	return document, nil
}

func selectCustomerReleaseRoot(
	path string,
	ackTrust []byte,
	now time.Time,
) (ed25519.PublicKey, error) {
	document, err := readCustomerAuthorityArtifact(path)
	if err != nil {
		return nil, err
	}
	defer clear(document)
	return platformauthority.SelectSaaSReleaseRoot(document, ackTrust, now)
}
