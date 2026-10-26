package iamactorv1

import (
	"net/http"
	"strings"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TakeAssertionHeader(header http.Header) (string, error) {
	if header == nil {
		return "", newError(CodeInvalid)
	}
	var values []string
	for name, candidates := range header {
		if strings.EqualFold(name, actorassertionv1.HeaderName) {
			values = append(values, candidates...)
			delete(header, name)
		}
	}
	if len(values) != 1 || values[0] == "" ||
		len(values[0]) > actorassertionv1.MaximumWireSize {
		return "", newError(CodeInvalid)
	}
	return values[0], nil
}

func (consumer *Consumer) ConsumeHeader(
	request *http.Request,
	transaction Transaction,
	expectation actorassertionv1.Expectation,
	effect Effect,
) (Receipt, error) {
	if request == nil {
		return Receipt{}, newError(CodeInvalid)
	}
	wire, err := TakeAssertionHeader(request.Header)
	if err != nil {
		return Receipt{}, err
	}
	return consumer.Consume(
		request.Context(), transaction, wire, expectation, effect,
	)
}
