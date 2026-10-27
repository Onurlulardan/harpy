package billingv1

import (
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) deleteCampaign(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-campaign-delete"
	access, err := handler.authorize(r, routeID, billingapi.PermissionPlatformBillingManage)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	canonicalID, found, err := handler.resolveDeleteIdentity(
		r, access, billingapi.LegacyIdentityCampaign, id,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if !found {
		writeCampaignDeleted(w)
		return
	}
	current, err := handler.catalog.GetCampaign(r.Context(), access, canonicalID)
	if errors.Is(err, billingapi.ErrNotFound) {
		writeCampaignDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if current.State == "retired" {
		writeCampaignDeleted(w)
		return
	}
	if current.State != "active" {
		handler.fail(w, routeID, ErrSnapshotRequired)
		return
	}
	key, err := handler.idempotencyKey(r)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	_, err = handler.catalog.RetireCampaign(r.Context(), billingapi.RetireCampaignCommand{
		Access: access, IdempotencyKey: key, CampaignID: current.ID,
		ExpectedRevision: current.Revision,
		Reason:           "legacy API delete mapped to immutable retirement",
	})
	if errors.Is(err, billingapi.ErrNotFound) {
		writeCampaignDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeCampaignDeleted(w)
}

func writeCampaignDeleted(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Campaign deleted"})
}
