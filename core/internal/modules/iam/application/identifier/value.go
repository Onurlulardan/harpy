package identifier

import "crypto/sha256"

type Value struct {
	kind       Kind
	display    string
	normalized string
}

func newValue(kind Kind, display, normalized string) Value {
	return Value{kind: kind, display: display, normalized: normalized}
}

func (value Value) Valid() bool {
	return value.kind.valid() && value.display != "" && value.normalized != ""
}

func (value Value) Kind() Kind {
	return value.kind
}

func (value Value) Display() string {
	return value.display
}

func (value Value) Normalized() string {
	return value.normalized
}

func (value Value) NormalizedSHA256() [sha256.Size]byte {
	return sha256.Sum256([]byte(value.normalized))
}
