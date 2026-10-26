package api

// GETDecoder decodes one canonical raw GET query into a query request.
type GETDecoder interface {
	Decode(rawQuery string) (QueryRequest, error)
}
