package ports

import "errors"

var (
	ErrProviderNotFound = errors.New("notification provider config was not found")
	ErrBinding          = errors.New("notification tenant store binding failed")
)
