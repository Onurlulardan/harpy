package airgap

import (
	"fmt"
	"os"
	"path/filepath"
)

const maximumArtifactBytes = 256 * 1024

func WriteRequestCreateOnly(path string, document []byte) error {
	if !filepath.IsAbs(path) || len(document) == 0 || len(document) > maximumArtifactBytes {
		return fmt.Errorf("air-gap request output is invalid")
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create air-gap request artifact: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(document); err != nil {
		return fmt.Errorf("write air-gap request artifact: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync air-gap request artifact: %w", err)
	}
	return nil
}

func ReadAcknowledgement(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("air-gap acknowledgement path is invalid")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 ||
		info.Size() > maximumArtifactBytes {
		return nil, fmt.Errorf("air-gap acknowledgement artifact is invalid")
	}
	document, err := os.ReadFile(path)
	if err != nil || len(document) != int(info.Size()) {
		return nil, fmt.Errorf("read air-gap acknowledgement artifact")
	}
	return document, nil
}
