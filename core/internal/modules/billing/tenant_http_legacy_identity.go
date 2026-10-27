package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

var _ billingapi.LegacyIdentityResolver = (*tenantOperations)(nil)

func (application *tenantOperations) ResolveLegacyIdentity(
	ctx context.Context,
	query billingapi.LegacyIdentityQuery,
) (billingapi.LegacyIdentityView, error) {
	identifier, err := uuid.Parse(query.LegacyID)
	if query.Kind != billingapi.LegacyIdentitySubscription || err != nil ||
		identifier == uuid.Nil || identifier.String() != query.LegacyID {
		return billingapi.LegacyIdentityView{}, billingapi.ErrInvalid
	}
	permission, valid := tenantIdentityPermission(query.Access)
	if !valid {
		return billingapi.LegacyIdentityView{}, billingapi.ErrForbidden
	}
	authority, err := application.resolve(ctx, query.Access, permission)
	if err != nil {
		return billingapi.LegacyIdentityView{}, err
	}
	snapshot, err := application.readSubscription(ctx, query.Access, authority)
	if err != nil {
		return billingapi.LegacyIdentityView{}, err
	}
	view := snapshot.View
	if view.LegacyID != query.LegacyID && view.ID != query.LegacyID {
		return billingapi.LegacyIdentityView{}, billingapi.ErrNotFound
	}
	canonical, err := uuid.Parse(view.ID)
	if err != nil || canonical.Version() != 7 {
		return billingapi.LegacyIdentityView{}, billingapi.ErrUnavailable
	}
	legacyID := view.LegacyID
	if legacyID == "" {
		legacyID = canonical.String()
	}
	return billingapi.LegacyIdentityView{
		LegacyID: legacyID, CanonicalID: canonical.String(),
	}, nil
}

func tenantIdentityPermission(access billingapi.Access) (billingapi.Permission, bool) {
	for _, permission := range []billingapi.Permission{
		billingapi.PermissionBillingRead,
		billingapi.PermissionBillingManage,
	} {
		if access.ValidFor(permission) {
			return permission, true
		}
	}
	return "", false
}
