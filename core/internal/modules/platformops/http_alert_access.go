package platformops

import "net/http"

func (handler *HTTPHandler) alertEnvironment(
	r *http.Request,
	capability string,
) (HTTPEnvironment, HTTPAccess, error) {
	environment, err := handler.environmentFor(r)
	if err != nil {
		return HTTPEnvironment{}, HTTPAccess{}, err
	}
	access, err := handler.accessFor(r, environment, capability, "")
	if err != nil {
		return HTTPEnvironment{}, HTTPAccess{}, err
	}
	return environment, access, nil
}
