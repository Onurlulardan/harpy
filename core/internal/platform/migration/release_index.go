package migration

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
)

const releaseIndexVersion = "harpy-release-index.v1"

type releaseIndex struct {
	ContractVersion string                 `json:"contract_version"`
	Digests         []string               `json:"digests"`
	Documents       []releaseIndexDocument `json:"documents"`
}

type releaseIndexDocument struct {
	Path            string `json:"path"`
	Owner           string `json:"owner"`
	Scope           string `json:"scope"`
	TransactionMode string `json:"transaction_mode"`
	SHA256          string `json:"sha256"`
}

func BuildReleaseIndex(files fs.FS, directory string) ([]byte, error) {
	registry, err := LoadReleaseManifestFS(files, directory)
	if err != nil {
		return nil, err
	}
	return encodeReleaseIndex(registry)
}

func encodeReleaseIndex(registry *ReleaseRegistry) ([]byte, error) {
	if registry == nil {
		return nil, fmt.Errorf("release index registry is unavailable")
	}
	documents := make([]releaseIndexDocument, 0, len(registry.documentContracts))
	paths := make([]string, 0, len(registry.documentContracts))
	for documentPath := range registry.documentContracts {
		paths = append(paths, documentPath)
	}
	sort.Strings(paths)
	for _, documentPath := range paths {
		contracts := make([]expectedDocument, 0, len(registry.documentContracts[documentPath]))
		for contract := range registry.documentContracts[documentPath] {
			contracts = append(contracts, contract)
		}
		sort.Slice(contracts, func(left, right int) bool {
			if contracts[left].owner != contracts[right].owner {
				return contracts[left].owner < contracts[right].owner
			}
			if contracts[left].scope != contracts[right].scope {
				return contracts[left].scope < contracts[right].scope
			}
			if contracts[left].transactionMode != contracts[right].transactionMode {
				return contracts[left].transactionMode < contracts[right].transactionMode
			}
			return contracts[left].sha256 < contracts[right].sha256
		})
		for _, contract := range contracts {
			documents = append(documents, releaseIndexDocument{
				Path: documentPath, Owner: contract.owner, Scope: contract.scope,
				TransactionMode: contract.transactionMode, SHA256: contract.sha256,
			})
		}
	}
	encoded, err := json.Marshal(releaseIndex{
		ContractVersion: releaseIndexVersion,
		Digests:         registry.Digests(),
		Documents:       documents,
	})
	if err != nil {
		return nil, fmt.Errorf("encode release index: %w", err)
	}
	return append(encoded, '\n'), nil
}

func parseReleaseIndex(document []byte) (releaseIndex, map[string]struct{}, map[string]map[expectedDocument]struct{}, error) {
	var index releaseIndex
	if err := json.Unmarshal(document, &index); err != nil {
		return releaseIndex{}, nil, nil, fmt.Errorf("decode release index: %w", err)
	}
	if index.ContractVersion != releaseIndexVersion || len(index.Digests) == 0 || len(index.Documents) == 0 {
		return releaseIndex{}, nil, nil, fmt.Errorf("release index is invalid")
	}
	known := make(map[string]struct{}, len(index.Digests))
	for _, digest := range index.Digests {
		if digest == "" {
			return releaseIndex{}, nil, nil, fmt.Errorf("release index digest is empty")
		}
		if _, duplicate := known[digest]; duplicate {
			return releaseIndex{}, nil, nil, fmt.Errorf("release index digest is duplicated")
		}
		known[digest] = struct{}{}
	}
	contracts := make(map[string]map[expectedDocument]struct{})
	for _, document := range index.Documents {
		if document.Path == "" || document.Owner == "" || document.Scope == "" ||
			document.TransactionMode == "" || document.SHA256 == "" {
			return releaseIndex{}, nil, nil, fmt.Errorf("release index document is invalid")
		}
		set := contracts[document.Path]
		if set == nil {
			set = map[expectedDocument]struct{}{}
			contracts[document.Path] = set
		}
		contract := expectedDocument{
			owner: document.Owner, scope: document.Scope,
			transactionMode: document.TransactionMode, sha256: document.SHA256,
		}
		if _, duplicate := set[contract]; duplicate {
			return releaseIndex{}, nil, nil, fmt.Errorf("release index document is duplicated")
		}
		set[contract] = struct{}{}
	}
	return index, known, contracts, nil
}
