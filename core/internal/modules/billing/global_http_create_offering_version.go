package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const offeringPublishOperation = "billing.catalog.offering.publish"

func (runtime *GlobalHTTPRuntime) CreateOfferingVersion(
	ctx context.Context, input billingapi.CreateOfferingVersionCommand,
) (billingapi.OfferingView, error) {
	command, readback, err := commandCapabilities(input.Access, offeringPublishOperation)
	if err != nil {
		return billingapi.OfferingView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.OfferingID)
	versionInput, inputErr := globalHTTPOfferingInput(input.Version)
	version, versionErr := offeringapp.BuildVersion(versionInput)
	if idErr != nil || inputErr != nil || versionErr != nil {
		return billingapi.OfferingView{}, billingapi.ErrInvalid
	}
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.offeringStore(scope)
			commands, buildErr := offeringapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			_, actionErr = commands.Publish(scoped, offeringapp.CreateVersionRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				OfferingID: identifier, ExpectedRevision: input.ExpectedRevision,
				Version: version,
			})
			return actionErr
		})
	if actionErr != nil {
		return billingapi.OfferingView{}, normalizeCommandError(actionErr)
	}
	if err != nil {
		return billingapi.OfferingView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readOffering(ctx, readback, identifier))
}
