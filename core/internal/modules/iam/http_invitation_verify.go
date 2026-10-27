package iam

import "net/http"

func (handler *InvitationHTTP) verify(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tokens, tenants := query["token"], query["tenant"]
	if len(query) != 2 || len(tokens) != 1 || len(tenants) != 1 ||
		tokens[0] == "" || tenants[0] == "" {
		handler.writeFailure(w, r, ErrInvitationHTTPInvalidCredential)
		return
	}
	token := []byte(tokens[0])
	tokens[0] = ""
	r.URL.RawQuery = ""
	r.RequestURI = r.URL.Path
	view, err := handler.application.Verify(
		r.Context(), InvitationCredentialInput{
			Token: token, TenantSlug: tenants[0],
		},
	)
	clear(token)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeJSON(w, r, http.StatusOK, view)
}
