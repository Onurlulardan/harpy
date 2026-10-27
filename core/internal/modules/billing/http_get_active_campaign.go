package billing

import "net/http"

func (handler *Handler) getActiveCampaign(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		handler.writeFailure(w, r, billingRequestInvalid)
		return
	}
	result, err := handler.public.GetActiveCampaign(r.Context())
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeSuccess(w, r, http.StatusOK, result)
}
