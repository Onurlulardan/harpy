package passwordblocklist

import (
	"crypto/sha256"
	"crypto/subtle"
	"io"
	"os"
	"path/filepath"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func Open(
	path string,
	expectedSHA256 [32]byte,
	kind blocklistv1.Kind,
) (*Blocklist, error) {
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path ||
		!kind.Valid() || allZero(expectedSHA256[:]) {
		return nil, ErrUnavailable
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, ErrUnavailable
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer file.Close()
	info, err := file.Stat()
	maximum := int64(16 + kind.MaximumEntries()*sha256.Size)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 16+sha256.Size ||
		info.Size() > maximum {
		return nil, ErrUnavailable
	}
	content, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || int64(len(content)) != info.Size() || int64(len(content)) > maximum {
		clear(content)
		return nil, ErrUnavailable
	}
	actual := sha256.Sum256(content)
	if subtle.ConstantTimeCompare(actual[:], expectedSHA256[:]) != 1 {
		clear(actual[:])
		clear(content)
		return nil, ErrUnavailable
	}
	clear(actual[:])
	artifact, err := blocklistv1.Parse(content, kind)
	if err != nil {
		clear(content)
		return nil, ErrUnavailable
	}
	return &Blocklist{content: content, artifact: artifact, open: true}, nil
}

func allZero(value []byte) bool {
	var combined byte
	for _, current := range value {
		combined |= current
	}
	return combined == 0
}
