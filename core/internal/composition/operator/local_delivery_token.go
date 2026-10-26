package operator

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func loadOrCreateToken(path string) ([]byte, string, error) {
	if _, err := os.Lstat(path); err == nil {
		content, readErr := localinstall.ReadRegular(path, 128)
		if readErr != nil {
			return nil, "", readErr
		}
		return tokenResult(content)
	} else if !os.IsNotExist(err) {
		return nil, "", err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		clear(raw)
		return nil, "", err
	}
	content := []byte(base64.RawURLEncoding.EncodeToString(raw))
	clear(raw)
	if err := localinstall.WriteExclusive(path, content, 0o600); err != nil {
		clear(content)
		return nil, "", err
	}
	return tokenResult(content)
}

func tokenResult(content []byte) ([]byte, string, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(string(content))
	if err != nil || len(decoded) != 32 {
		clear(content)
		clear(decoded)
		return nil, "", installationv1Error{}
	}
	clear(decoded)
	digest := sha256.Sum256(content)
	return content, hex.EncodeToString(digest[:]), nil
}
