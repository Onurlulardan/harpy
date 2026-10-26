package asyncmanifestv1

import eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"

const (
	AnalyticsQueryPublishedEventType          = "analytics.query.published"
	AnalyticsQueryPublishedSchemaSHA256       = "a6468a0f99cf339c986ad58985e51a7d0e3be6ebee830b2e62c19b3a0c5babeb"
	AnalyticsDashboardPublishedEventType      = "analytics.dashboard.published"
	AnalyticsDashboardPublishedSchemaSHA256   = "e9cf0712d0be6f24508f5560d083a16a0cac889940ba290b4e66dfb16f901934"
	AnalyticsPublicationActivatedEventType    = "analytics.publication.activated"
	AnalyticsPublicationActivatedSchemaSHA256 = "0a415d8297153e8d533fba85ebab567198a5a353fb5720c1c00543e29ac71a5d"
	AnalyticsPublicationRevokedEventType      = "analytics.publication.revoked"
	AnalyticsPublicationRevokedSchemaSHA256   = "0f6b3771ef2798bb9a2ce17c34aa5a7e5b31471fab81f66909c1f2956a045cd0"
)

func AnalyticsProducerPolicies() []eventsv1.ProducerPolicy {
	definitions := [][2]string{
		{AnalyticsQueryPublishedEventType, AnalyticsQueryPublishedSchemaSHA256},
		{AnalyticsDashboardPublishedEventType, AnalyticsDashboardPublishedSchemaSHA256},
		{AnalyticsPublicationActivatedEventType, AnalyticsPublicationActivatedSchemaSHA256},
		{AnalyticsPublicationRevokedEventType, AnalyticsPublicationRevokedSchemaSHA256},
	}
	policies := make([]eventsv1.ProducerPolicy, len(definitions))
	for index, definition := range definitions {
		policies[index], _ = eventsv1.NewProducerPolicy(definition[0], 1, "analytics", definition[1])
	}
	return policies
}
