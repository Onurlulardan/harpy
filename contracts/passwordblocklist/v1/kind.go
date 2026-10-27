package passwordblocklistv1

const (
	headerSize = 16
	digestSize = 32
)

var artifactMagic = [8]byte{'F', 'C', 'B', 'L', 'S', 'T', '0', '1'}

type Kind uint8

const (
	KindCommon      Kind = 1
	KindCompromised Kind = 2
)

func (kind Kind) Valid() bool {
	return kind == KindCommon || kind == KindCompromised
}

func (kind Kind) MaximumEntries() int {
	switch kind {
	case KindCommon:
		return 500_000
	case KindCompromised:
		return 8_000_000
	default:
		return 0
	}
}

func (kind Kind) String() string {
	switch kind {
	case KindCommon:
		return "common"
	case KindCompromised:
		return "compromised"
	default:
		return "invalid"
	}
}
