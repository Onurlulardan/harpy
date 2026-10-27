package api

// POSTDecoder decodes the canonical POST wire document into a query request.
type POSTDecoder interface {
	Decode(document []byte) (QueryRequest, error)
}
