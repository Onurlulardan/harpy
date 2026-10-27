package billingv1

import (
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) deletePlan(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-plan-delete"
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
	identity, err := handler.resolvePlanIdentity(r.Context(), access, id)
	if errors.Is(err, billingapi.ErrNotFound) {
		writePlanDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	current, err := handler.catalog.GetOffering(
		r.Context(), access, identity.CanonicalID,
	)
	if errors.Is(err, billingapi.ErrNotFound) {
		writePlanDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if current.State == "retired" {
		writePlanDeleted(w)
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
	_, err = handler.catalog.RetireOffering(r.Context(), billingapi.RetireOfferingCommand{
		Access: access, IdempotencyKey: key, OfferingID: current.ID,
		ExpectedRevision: current.Revision,
		Reason:           "legacy API delete mapped to immutable retirement",
	})
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writePlanDeleted(w)
}

func writePlanDeleted(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Plan deleted"})
}
