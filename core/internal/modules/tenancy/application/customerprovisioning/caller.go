package customerprovisioning

type CallerKind string

const (
	CallerHTTP     CallerKind = "http"
	CallerInternal CallerKind = "internal"
)

func (kind CallerKind) valid() bool {
	return kind == CallerHTTP || kind == CallerInternal
}
