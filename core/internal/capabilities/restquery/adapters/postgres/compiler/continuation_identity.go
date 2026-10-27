package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func continuationIdentity(
	query resolved.RecordQuery,
) ([sha256.Size]byte, error) {
	page, err := query.Page()
	if err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	token, present, err := page.After()
	if err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	writer := newSealWriter(
		"forgecrud.restquery.postgres.execution-continuation.v1",
	)
	writer.writeBool(present)
	if present {
		writer.writeString(token)
	}
	return writer.sum()
}
