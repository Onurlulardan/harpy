package checkpoint

type ScopeLeaf struct {
	RegistryBindingID string `json:"registry_binding_id"`
	ScopeKind         string `json:"scope_kind"`
	ScopeID           string `json:"scope_id"`
	Generation        int64  `json:"generation"`
	StateSHA256       string `json:"state_sha256"`
	AckSHA256         string `json:"ack_sha256"`
}

type InclusionProof struct {
	Leaf      ScopeLeaf
	LeafIndex int
	LeafCount int
	Siblings  []string
}

type Tree struct {
	RootSHA256 string
	Leaves     []ScopeLeaf
	proofs     []InclusionProof
}
