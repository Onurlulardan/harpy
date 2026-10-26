package operator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
)

const maximumInitialPasswordBytes = 1024

func readInstallationManifest(path config.ArtifactPath) (installationmanifest.Manifest, error) {
	document, err := readRegularFile(path.Path().Value(), installationmanifest.MaximumBytes, false)
	if err != nil {
		return installationmanifest.Manifest{}, err
	}
	return installationmanifest.Parse(document)
}

func passwordLoader(path config.ArtifactPath) InitialCredentialLoader {
	return func(_ context.Context) ([]byte, error) {
		content, err := readRegularFile(path.Path().Value(), maximumInitialPasswordBytes, true)
		if err != nil {
			return nil, err
		}
		content = removeOneTerminalNewline(content)
		if len(content) == 0 || bytes.IndexByte(content, 0) >= 0 {
			clear(content)
			return nil, fmt.Errorf("initial password file is invalid")
		}
		return content, nil
	}
}

func readRegularFile(path string, maximum int, ownerOnly bool) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 ||
		before.Size() < 1 || before.Size() > int64(maximum) ||
		ownerOnly && runtime.GOOS != "windows" && before.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("operator artifact is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("operator artifact is unavailable")
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return nil, fmt.Errorf("operator artifact changed while opening")
	}
	content, err := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
	if err != nil || len(content) > maximum {
		clear(content)
		return nil, fmt.Errorf("operator artifact exceeds its limit")
	}
	return content, nil
}

func removeOneTerminalNewline(content []byte) []byte {
	if bytes.HasSuffix(content, []byte("\r\n")) {
		return content[:len(content)-2]
	}
	if bytes.HasSuffix(content, []byte("\n")) {
		return content[:len(content)-1]
	}
	return content
}
