package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const (
	completeMarker = ".projection.json"
	pendingMarker  = ".projection.pending.json"
)

type projectionMarker struct {
	Version      int          `json:"version"`
	GenerationID string       `json:"generation_id"`
	Target       string       `json:"target"`
	Files        []sourceFile `json:"files"`
}

func markerFor(source verifiedSource, target targetSpec) projectionMarker {
	marker := projectionMarker{
		Version: 1, GenerationID: source.generation, Target: target.name,
		Files: make([]sourceFile, 0, len(target.files)),
	}
	for _, file := range target.files {
		marker.Files = append(marker.Files, source.files[file.name])
	}
	sort.Slice(marker.Files, func(left, right int) bool {
		return marker.Files[left].Name < marker.Files[right].Name
	})
	return marker
}

func marshalMarker(marker projectionMarker) ([]byte, error) {
	content, err := json.Marshal(marker)
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

func decodeMarker(content []byte) (projectionMarker, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var marker projectionMarker
	if err := decoder.Decode(&marker); err != nil {
		return projectionMarker{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return projectionMarker{}, fmt.Errorf("projection marker has trailing content")
	}
	return marker, nil
}

func markersEqual(left, right projectionMarker) bool {
	if left.Version != right.Version || left.GenerationID != right.GenerationID ||
		left.Target != right.Target || len(left.Files) != len(right.Files) {
		return false
	}
	for index := range left.Files {
		if left.Files[index] != right.Files[index] {
			return false
		}
	}
	return true
}
