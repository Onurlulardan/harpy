package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeManifest(root string, replacements []generatedFile) error {
	path := filepath.Join(root, "contracts", "generated.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	updated, err := replaceGeneratedRecords(data, replacements)
	if err != nil {
		return err
	}
	if bytes.Equal(data, updated) {
		return nil
	}
	return os.WriteFile(path, updated, 0o644)
}

func replaceGeneratedRecords(data []byte, replacements []generatedFile) ([]byte, error) {
	raws, err := rawFilesByPath(data)
	if err != nil {
		return nil, err
	}
	out := data
	for _, replacement := range replacements {
		encoded, err := marshalRecord(replacement, data)
		if err != nil {
			return nil, err
		}
		current, ok := raws[replacement.Path]
		if !ok {
			out, err = insertRecord(out, encoded)
			if err != nil {
				return nil, err
			}
			continue
		}
		out, err = spliceRecord(out, current, encoded)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func rawFilesByPath(data []byte) (map[string][]byte, error) {
	var manifest struct {
		Files []json.RawMessage `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest.Files == nil {
		return nil, fmt.Errorf("generated manifest files must be an array")
	}
	raws := make(map[string][]byte, len(manifest.Files))
	for _, file := range manifest.Files {
		var entry generatedFile
		if err := json.Unmarshal(file, &entry); err != nil {
			return nil, err
		}
		if _, exists := raws[entry.Path]; exists {
			return nil, fmt.Errorf("duplicate generated entry %s", entry.Path)
		}
		raws[entry.Path] = append([]byte(nil), file...)
	}
	return raws, nil
}

func marshalRecord(file generatedFile, source []byte) ([]byte, error) {
	encoded, err := json.MarshalIndent(file, "    ", "  ")
	if err != nil {
		return nil, err
	}
	if bytes.Contains(source, []byte("\r\n")) {
		encoded = bytes.ReplaceAll(encoded, []byte("\n"), []byte("\r\n"))
	}
	return encoded, nil
}

func spliceRecord(data, current, encoded []byte) ([]byte, error) {
	index := bytes.Index(data, current)
	if index < 0 || bytes.Contains(data[index+len(current):], current) {
		return nil, fmt.Errorf("currency manifest record is not a unique span")
	}
	out := make([]byte, 0, len(data)-len(current)+len(encoded))
	out = append(out, data[:index]...)
	out = append(out, encoded...)
	out = append(out, data[index+len(current):]...)
	return out, nil
}

func insertRecord(data, encoded []byte) ([]byte, error) {
	var manifest struct {
		Files []json.RawMessage `json:"files"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if len(manifest.Files) == 0 {
		return nil, fmt.Errorf("generated manifest has no files array to extend")
	}
	last := manifest.Files[len(manifest.Files)-1]
	index := bytes.LastIndex(data, last)
	if index < 0 {
		return nil, fmt.Errorf("generated manifest files array is not addressable")
	}
	newline := []byte("\n")
	if bytes.Contains(data, []byte("\r\n")) {
		newline = []byte("\r\n")
	}
	piece := append(append([]byte(","), newline...), append([]byte("    "), encoded...)...)
	at := index + len(last)
	out := make([]byte, 0, len(data)+len(piece))
	out = append(out, data[:at]...)
	out = append(out, piece...)
	out = append(out, data[at:]...)
	return out, nil
}
