package billingv1

import (
	"errors"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) deleteCoupon(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.admin-coupon-delete"
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
		r, access, billingapi.LegacyIdentityDiscount, id,
	)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if !found {
		writeCouponDeleted(w)
		return
	}
	current, err := handler.catalog.GetDiscount(r.Context(), access, canonicalID)
	if errors.Is(err, billingapi.ErrNotFound) {
		writeCouponDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	if current.State == "retired" {
		writeCouponDeleted(w)
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
	_, err = handler.catalog.RetireDiscount(r.Context(), billingapi.RetireDiscountCommand{
		Access: access, IdempotencyKey: key, DiscountID: current.ID,
		ExpectedRevision: current.Revision,
		Reason:           "legacy API delete mapped to immutable retirement",
	})
	if errors.Is(err, billingapi.ErrNotFound) {
		writeCouponDeleted(w)
		return
	}
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeCouponDeleted(w)
}

func writeCouponDeleted(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Coupon deleted"})
}
