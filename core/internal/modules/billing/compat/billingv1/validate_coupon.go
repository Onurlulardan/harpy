package billingv1

import (
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

type validateCouponRequest struct {
	Code   string `json:"code"`
	PlanID string `json:"plan_id"`
}

func (handler *Handler) validateCoupon(w http.ResponseWriter, r *http.Request) {
	const routeID = "billing.route.validate-coupon"
	access, err := handler.authorize(r, routeID, billingapi.PermissionBillingRead)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	input, err := decodeBody[validateCouponRequest](r)
	if err != nil || !required(input.Code) {
		handler.fail(w, routeID, ErrRequestInvalid)
		return
	}
	offeringID, err := handler.validationOfferingID(r, input.PlanID)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	view, err := handler.tenant.ValidateDiscount(r.Context(), billingapi.ValidateDiscountCommand{
		Access: access, Code: input.Code, OfferingVersionID: offeringID,
	})
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	wire, err := legacyCoupon(view)
	if err != nil {
		handler.fail(w, routeID, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"valid": true, "code": wire.Code, "name": wire.Name,
		"discount_type": wire.DiscountType, "discount_value": wire.DiscountValue,
		"duration": wire.Duration,
	})
}

func (handler *Handler) validationOfferingID(
	r *http.Request,
	planID string,
) (string, error) {
	if planID == "" || zeroUUID(planID) {
		return "", nil
	}
	identity, err := handler.resolvePublicPlanIdentity(r.Context(), planID)
	if err != nil {
		return "", err
	}
	offerings, err := handler.publicOfferings(r.Context())
	if err != nil {
		return "", err
	}
	offering, err := findOffering(offerings, identity.CanonicalID)
	return offering.VersionID, err
}

func zeroUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed == uuid.Nil
}
