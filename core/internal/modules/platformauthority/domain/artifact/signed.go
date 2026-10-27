package artifact

type SignedProposal struct {
	Value  Proposal
	Bytes  []byte
	SHA256 string
}

type VerifiedAck struct {
	Value  Ack
	Bytes  []byte
	SHA256 string
}

type VerifiedBindingProof struct {
	Value  BindingProof
	Bytes  []byte
	SHA256 string
}

type VerifiedCheckpoint struct {
	Value  Checkpoint
	Bytes  []byte
	SHA256 string
}

type ExpiryPolicy string

const (
	ExpiryOnline ExpiryPolicy = "online"
	ExpiryAirGap ExpiryPolicy = "airgap"
)
