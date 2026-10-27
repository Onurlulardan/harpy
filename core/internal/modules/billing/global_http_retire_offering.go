package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const offeringRetireOperation = "billing.catalog.offering.retire"

func (runtime *GlobalHTTPRuntime) RetireOffering(
	ctx context.Context, input billingapi.RetireOfferingCommand,
) (billingapi.OfferingView, error) {
	command, readback, err := commandCapabilities(input.Access, offeringRetireOperation)
	if err != nil {
		return billingapi.OfferingView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.OfferingID)
	if idErr != nil {
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
			_, actionErr = commands.Retire(scoped, offeringapp.RetireRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				OfferingID: identifier, ExpectedRevision: input.ExpectedRevision,
				Reason: input.Reason,
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
