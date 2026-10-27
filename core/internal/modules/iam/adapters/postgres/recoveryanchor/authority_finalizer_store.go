package recoveryanchorpostgres

import "context"

type authorityFinalizerScanner interface {
	Scan(...any) error
}

type authorityFinalizerDatabase interface {
	QueryRow(context.Context, string, ...any) authorityFinalizerScanner
}

type authorityFinalizerTransactions interface {
	Within(
		context.Context,
		func(authorityFinalizerDatabase) error,
	) error
}

type authorityFinalizerStore struct {
	transactions authorityFinalizerTransactions
}
