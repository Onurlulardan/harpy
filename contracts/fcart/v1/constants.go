package fcartv1

const (
	Version               = "FCART1"
	ChunkPlaintextBytes   = 8 * 1024 * 1024
	MaxCiphertextBytes    = uint64(2_147_483_648)
	DEKBytes              = 32
	NonceBytes            = 12
	TagBytes              = 16
	HeaderBytes           = len(Version) + 16 + 8 + 1 + 4
	ChunkFramePrefixBytes = 8 + 4 + NonceBytes
	ChunkFramingBytes     = ChunkFramePrefixBytes + TagBytes
)
