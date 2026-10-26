package platformops

import "net/http"

func (handler *HTTPHandler) traceEnvironment(
	r *http.Request,
) (HTTPEnvironment, error) {
	environment, err := handler.environmentFor(r)
	if err != nil {
		return HTTPEnvironment{}, err
	}
	if _, err = handler.accessFor(
		r, environment, "platform.operations.traces.read", "",
	); err != nil {
		return HTTPEnvironment{}, err
	}
	return environment, nil
}
