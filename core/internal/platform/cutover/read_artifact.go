package cutover

import (
	"fmt"
	"io"
	"os"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

const (
	maximumCutoverAuthorityBytes = 16 * 1024 * 1024
	maximumCutoverBatchBytes     = cutoverv1.MaximumChangeBatchBytes
	maximumCutoverReceiptBytes   = 64 * 1024
)

func readArtifact(path string, maximum int64) ([]byte, error) {
	if path == "" || maximum < 1 {
		return nil, fmt.Errorf("cutover artifact input is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open cutover artifact")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maximum {
		return nil, fmt.Errorf("cutover artifact is not a bounded regular file")
	}
	document, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || len(document) == 0 || int64(len(document)) > maximum {
		return nil, fmt.Errorf("read bounded cutover artifact")
	}
	return document, nil
}
