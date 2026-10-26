package identifier

type Kind string

const (
	EmailV1  Kind = "email_v1"
	PortalV1 Kind = "portal_v1"
)

func (kind Kind) valid() bool {
	return kind == EmailV1 || kind == PortalV1
}
