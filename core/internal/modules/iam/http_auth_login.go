package iam

import (
	"net/http"
)

func (handler *AuthHTTP) login(w http.ResponseWriter, r *http.Request) {
	input, err := decodeOperatorLogin(w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	handler.executeLogin(w, r, input)
}
