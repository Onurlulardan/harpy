package deployevidence

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type artifactDocument struct {
	bytes        []byte
	resolvedPath string
}

func readArtifact(path string, maximumBytes int64, label string) (artifactDocument, error) {
	if path == "" || maximumBytes < 1 || label == "" {
		return artifactDocument{}, fmt.Errorf("artifact input is invalid")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return artifactDocument{}, fmt.Errorf("%s artifact is unavailable", label)
	}
	file, err := os.Open(resolved)
	if err != nil {
		return artifactDocument{}, fmt.Errorf("%s artifact is unavailable", label)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maximumBytes {
		return artifactDocument{}, fmt.Errorf("%s artifact is not a bounded regular file", label)
	}
	content, err := io.ReadAll(io.LimitReader(file, maximumBytes+1))
	if err != nil || int64(len(content)) > maximumBytes {
		return artifactDocument{}, fmt.Errorf("%s artifact could not be read", label)
	}
	return artifactDocument{bytes: content, resolvedPath: resolved}, nil
}
