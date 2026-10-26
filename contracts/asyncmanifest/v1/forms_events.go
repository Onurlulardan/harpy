package asyncmanifestv1

import eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"

const (
	FormsPublicationCompletedEventType    = "forms.publication.completed"
	FormsPublicationCompletedSchemaSHA256 = "977788cec22783f146663729fc11a9ac39a3eb6374b46f7c77a35ef3b519b90f"
	FormsRecordCreatedEventType           = "forms.record.created"
	FormsRecordCreatedSchemaSHA256        = "a81e18d2434a88aff6b58b9a8bc19b6d33e612eae118c259ea88b269fb2835d4"
	FormsRecordDeletedEventType           = "forms.record.deleted"
	FormsRecordDeletedSchemaSHA256        = "833ca2a4a4ee25014b4b1c4d1f19f8f386829ffdf18931a5aa52894f43bcca98"
	FormsRecordUpdatedEventType           = "forms.record.updated"
	FormsRecordUpdatedSchemaSHA256        = "3c6c3b4d5b3506f417c609ed3896decdf964abe70661119cdf77181f75a74ff2"
)

func FormsProducerPolicies() []eventsv1.ProducerPolicy {
	definitions := []struct{ eventType, schemaSHA256 string }{
		{FormsPublicationCompletedEventType, FormsPublicationCompletedSchemaSHA256},
		{FormsRecordCreatedEventType, FormsRecordCreatedSchemaSHA256},
		{FormsRecordDeletedEventType, FormsRecordDeletedSchemaSHA256},
		{FormsRecordUpdatedEventType, FormsRecordUpdatedSchemaSHA256},
	}
	policies := make([]eventsv1.ProducerPolicy, len(definitions))
	for index, definition := range definitions {
		policies[index], _ = eventsv1.NewProducerPolicy(
			definition.eventType, 1, "forms", definition.schemaSHA256,
		)
	}
	return policies
}
