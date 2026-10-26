package main

import (
	"net/http"
	"strings"

	"google.golang.org/grpc"
)

func newHTTPHandler(
	grpcServer *grpc.Server,
	authorities ...*tokenAuthority,
) http.Handler {
	authority := localTokenAuthority()
	if len(authorities) == 1 && authorities[0] != nil {
		authority = authorities[0]
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.ProtoMajor == 2 && strings.HasPrefix(
			request.Header.Get("Content-Type"), "application/grpc",
		) {
			grpcServer.ServeHTTP(response, request)
			return
		}
		switch request.URL.Path {
		case "/v1/token":
			tokenExchange(response, request, authority)
		case "/livez":
			if request.Method != http.MethodGet {
				response.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			response.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(response, request)
		}
	})
}
