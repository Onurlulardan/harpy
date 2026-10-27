package operator

import (
	"context"
	"time"

	licensebootstrap "github.com/Onurlulardan/harpy/core/internal/capabilities/licensebootstrap/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func publishOffering(
	ctx context.Context,
	database *billingDatabase,
	trust billing.LocalLicenseTrustBundle,
	snapshot billing.LocalLicenseVerifiedSnapshot,
	databaseNow time.Time,
	identity localinstall.Identity,
) error {
	plan := billing.LocalLicenseOfferingPlan{
		OfferingVersionID: identity.OfferingVersionID,
		Revision:          1,
		EffectiveAt:       snapshot.NotBefore(),
	}
	document, err := billing.BuildLocalLicenseOfferingDocument(plan, snapshot)
	if err != nil {
		return err
	}
	store, err := billing.NewLocalLicenseOfferingStore(
		func(
			ctx context.Context, statement string, arguments ...any,
		) billing.LocalLicenseDatabaseRow {
			return database.inner.QueryRow(ctx, statement, arguments...)
		},
	)
	if err != nil {
		return err
	}
	checker, err := billing.NewLocalLicenseSnapshotChecker(trust)
	if err != nil {
		return err
	}
	publisher, err := billing.NewLocalLicenseOfferingPublisher(store, checker)
	if err != nil {
		return err
	}
	binding := licensebootstrap.Binding{
		OperationID:   identity.OfferingOperationID,
		EnvironmentID: snapshot.EnvironmentID(), LicenseID: snapshot.LicenseID(),
		LicenseeOrganizationID: snapshot.OrganizationID(), DeploymentMode: "onpremise",
		DeliveryKind: string(snapshot.DeliveryKind()), Issuer: snapshot.Issuer(),
		Sequence: snapshot.Sequence(), SnapshotSHA256: snapshot.Digest(),
		OfferingVersionID: identity.OfferingVersionID, Revision: 1,
		OfferingSHA256: document.SHA256, EffectiveAt: plan.EffectiveAt,
		ManifestSHA256: billing.LocalLicenseBootstrapManifestSHA256(),
	}
	grant := licensebootstrap.Grant{
		Binding: binding, Revision: 1, LeaseUntil: databaseNow.Add(time.Minute),
	}
	return licensebootstrap.Run(grant, databaseNow, func(scope licensebootstrap.Scope) error {
		_, publishErr := publisher.Publish(
			ctx, scope, snapshot.SignedDocument(), databaseNow)
		return publishErr
	})
}
