package archive

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type requestWire struct {
	ItemIDs []string `json:"item_ids"`
	RetryOf string   `json:"retry_of,omitempty"`
}

func canonicalRequest(request CreateRequest) ([]domain.ID, []byte, [32]byte, bool) {
	if len(request.ItemIDs) == 0 || len(request.ItemIDs) > domain.MaximumArchiveFiles {
		return nil, nil, [32]byte{}, false
	}
	ids := append([]domain.ID(nil), request.ItemIDs...)
	sort.Slice(ids, func(left, right int) bool {
		return bytes.Compare(ids[left][:], ids[right][:]) < 0
	})
	wire := requestWire{ItemIDs: make([]string, len(ids))}
	for index, id := range ids {
		if !id.UUIDv7() || index > 0 && id == ids[index-1] {
			return nil, nil, [32]byte{}, false
		}
		wire.ItemIDs[index] = id.String()
	}
	if request.RetryOfArchiveID.Valid() {
		if !request.RetryOfArchiveID.UUIDv7() {
			return nil, nil, [32]byte{}, false
		}
		wire.RetryOf = request.RetryOfArchiveID.String()
	}
	document, err := json.Marshal(wire)
	if err != nil || len(document) > 65536 {
		return nil, nil, [32]byte{}, false
	}
	return ids, document, sha256.Sum256(document), true
}

func jobDocument(archiveID domain.ID) []byte {
	value, _ := json.Marshal(struct {
		ArchiveID string `json:"archive_id"`
	}{ArchiveID: archiveID.String()})
	return value
}
