package policy

import (
	"crypto/sha256"
	"encoding/json"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

type revisionCatalog struct {
	ResourceID string `json:"resource_id"`
	VersionID  string `json:"version_id"`
	Hash       string `json:"hash"`
}

type revisionDocument struct {
	ContractVersion      string            `json:"contract_version"`
	ActorID              string            `json:"actor_id"`
	TokenVersion         int64             `json:"token_version"`
	AuthorizationVersion int64             `json:"authorization_version"`
	AuthorizationClock   int64             `json:"authorization_clock"`
	Catalogs             []revisionCatalog `json:"catalogs"`
}

func contextRevision(state *evaluationState) ([sha256.Size]byte, error) {
	if state == nil || !state.clockSet || state.versions.Validate() != nil ||
		len(state.catalogs.values) == 0 {
		return [sha256.Size]byte{}, catalog.ErrContract
	}
	entries := make([]revisionCatalog, 0, len(state.catalogs.values))
	for _, published := range state.catalogs.values {
		reference, err := published.Reference()
		if err != nil {
			return [sha256.Size]byte{}, catalog.ErrContract
		}
		resourceID, resourceErr := reference.ResourceID()
		versionID, versionErr := reference.PublishedVersionID()
		hash, hashErr := reference.CatalogHash()
		hashHex, hexErr := hash.Hex()
		if resourceErr != nil || versionErr != nil || hashErr != nil || hexErr != nil {
			return [sha256.Size]byte{}, catalog.ErrContract
		}
		entries = append(entries, revisionCatalog{
			ResourceID: resourceID.String(), VersionID: versionID.String(), Hash: hashHex,
		})
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].ResourceID < entries[right].ResourceID
	})
	document := revisionDocument{
		ContractVersion:      "iam-rest-query-policy.v1",
		ActorID:              state.versions.ActorID().String(),
		TokenVersion:         state.versions.TokenVersion(),
		AuthorizationVersion: state.versions.AuthorizationVersion(),
		AuthorizationClock:   state.clock, Catalogs: entries,
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return [sha256.Size]byte{}, catalog.ErrContract
	}
	digest := sha256.Sum256(canonical)
	if digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, catalog.ErrContract
	}
	return digest, nil
}
