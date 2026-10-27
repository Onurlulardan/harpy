package smtp

import (
	"context"
	"time"
)

type exchangeRequest struct {
	host, security, from, target string
	port                         int
	username, password, message  []byte
}

type exchangeResult struct {
	accepted, definitive, temporary, bytesWritten bool
	code, class, statusCode                       string
	completedAt                                   time.Time
}

type exchanger interface {
	Exchange(context.Context, exchangeRequest) exchangeResult
}
