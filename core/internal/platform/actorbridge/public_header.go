package actorbridge

import (
	"net/http"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func RejectPublicActorAssertion(next http.Handler) (http.Handler, error) {
	if next == nil {
		return nil, NewFault(FaultInvalid)
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request == nil || hasHeaderFold(request.Header, actorassertionv1.HeaderName) {
			writer.Header().Set("Connection", "close")
			http.Error(writer, "invalid request", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(writer, request)
	}), nil
}
