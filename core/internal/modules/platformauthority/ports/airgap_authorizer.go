package ports

import "context"

type AirGapAuthorizer interface {
	AuthorizeAckImport(context.Context, string, string) error
}
