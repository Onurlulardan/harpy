package iam

import "net/http"

func (handler *AuthHTTP) portalLogin(w http.ResponseWriter, r *http.Request) {
	input, err := decodePortalLogin(w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	handler.executeLogin(w, r, input)
}
