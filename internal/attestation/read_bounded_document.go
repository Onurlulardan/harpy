package attestation

import (
	"fmt"
	"io"
)

func readBoundedDocument(reader io.Reader, maximumBytes int64) ([]byte, error) {
	if reader == nil || maximumBytes < 1 {
		return nil, fmt.Errorf("attestation reader bound is invalid")
	}
	document, err := io.ReadAll(io.LimitReader(reader, maximumBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read attestation document: %w", err)
	}
	if len(document) == 0 || int64(len(document)) > maximumBytes {
		return nil, fmt.Errorf("attestation document size is invalid")
	}
	return document, nil
}
