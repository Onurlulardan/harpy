package sqlidentifier

type tokenKind uint8

const (
	tokenKindStatic tokenKind = iota + 1
	tokenKindDynamic
)

// Token is an opaque, validated PostgreSQL identifier capability.
type Token struct {
	identifier string
	kind       tokenKind
}
