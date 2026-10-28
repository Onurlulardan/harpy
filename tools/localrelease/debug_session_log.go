package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func agentDebugLog(hypothesisID, location, message string, data map[string]string) {
	// #region agent log
	line, err := json.Marshal(map[string]any{
		"sessionId":    "66ab43",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
		"runId":        "pre-fix",
	})
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "debug %s %s\n", hypothesisID, message)
	for _, path := range debugLogPaths() {
		file, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if openErr != nil {
			continue
		}
		_, _ = file.Write(append(line, '\n'))
		_ = file.Close()
	}
	// #endregion
}

func debugLogPaths() []string {
	paths := []string{"/tmp/debug-66ab43.log"}
	dir, err := os.Getwd()
	if err != nil {
		return paths
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return append(paths, filepath.Join(dir, "debug-66ab43.log"))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return paths
		}
		dir = parent
	}
}
