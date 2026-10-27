package webhookbody

import "context"

type Wrapper interface {
	Wrap(context.Context, []byte) (WrappedDEK, error)
}

type Unwrapper interface {
	Unwrap(context.Context, WrappedDEK) ([]byte, error)
}
