package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const offeringCreateOperation = "billing.catalog.offering.create"

func (runtime *GlobalHTTPRuntime) CreateOffering(
	ctx context.Context, input billingapi.CreateOfferingCommand,
) (billingapi.OfferingView, error) {
	command, readback, err := commandCapabilities(input.Access, offeringCreateOperation)
	if err != nil {
		return billingapi.OfferingView{}, err
	}
	key, keyErr := offeringapp.ParseKey(input.Key)
	versionInput, inputErr := globalHTTPOfferingInput(input.InitialVersion)
	version, versionErr := offeringapp.BuildVersion(versionInput)
	if keyErr != nil || inputErr != nil || versionErr != nil {
		return billingapi.OfferingView{}, billingapi.ErrInvalid
	}
	var identifier uuid.UUID
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.offeringStore(scope)
			commands, buildErr := offeringapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			result, executeErr := commands.Create(scoped, offeringapp.CreateRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				Key: key, Version: version,
			})
			actionErr, identifier = executeErr, result.OfferingID
			return actionErr
		})
	if actionErr != nil {
		return billingapi.OfferingView{}, normalizeCommandError(actionErr)
	}
	if err != nil || identifier == uuid.Nil {
		return billingapi.OfferingView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readOffering(ctx, readback, identifier))
}
