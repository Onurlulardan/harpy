package asyncmanifestv1

import eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"

const (
	CatalogInstallationActivatedEventType = "catalog.installation.activated"
	CatalogInstallationDisabledEventType  = "catalog.installation.disabled"
	CatalogInstallationRetiredEventType   = "catalog.installation.retired"
	CatalogInstallationPurgedEventType    = "catalog.installation.purged"
	CatalogMenuPublishedEventType         = "catalog.menu.published"
	CatalogLifecycleSchemaSHA256          = "49ef092d648660fde1bd02f6ffeb75522e891032432645d14e4bfcbf2f7aad3a"
	CatalogMenuSchemaSHA256               = "170b6496c75165bd88a57222ce3193f51daa48e66e44bc984430788bc4166d3d"
)

func CatalogProducerPolicies() []eventsv1.ProducerPolicy {
	definitions := [][2]string{
		{CatalogInstallationActivatedEventType, CatalogLifecycleSchemaSHA256},
		{CatalogInstallationDisabledEventType, CatalogLifecycleSchemaSHA256},
		{CatalogInstallationRetiredEventType, CatalogLifecycleSchemaSHA256},
		{CatalogInstallationPurgedEventType, CatalogLifecycleSchemaSHA256},
		{CatalogMenuPublishedEventType, CatalogMenuSchemaSHA256},
	}
	policies := make([]eventsv1.ProducerPolicy, len(definitions))
	for index, definition := range definitions {
		policies[index], _ = eventsv1.NewProducerPolicy(definition[0], 1, "catalog", definition[1])
	}
	return policies
}
