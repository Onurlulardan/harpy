package eventsv1

const (
	EnvelopeContractVersion        = "integration-event.v1"
	ProducerPolicyContractVersion  = "event-producer-policy.v1"
	ConsumerPolicyContractVersion  = "event-consumer-policy.v1"
	BootstrapContractVersion       = "event-bootstrap-receipt.v1"
	EffectReceiptContractVersion   = "event-effect-receipt.v1"
	OperatorReceiptContractVersion = "event-operator-receipt.v1"
	SequenceGapContractVersion     = "event-sequence-gap.v1"

	MinimumDeliveryLimit            = 1
	MaximumDeliveryLimit            = 25
	DeliveredOutboxRetentionDays    = 30
	UndeliveredOutboxAlarmSeconds   = 5 * 60
	InboxRetentionDays              = 180
	DeadLetterFullPayloadDays       = 30
	DeadLetterRedactedRetentionDays = 365
	BrokerRedeliveryHorizonDays     = 30
	MaximumCanonicalPayloadBytes    = 65_536

	MaximumDomainBytes      = 63
	MaximumTypeBytes        = 128
	MaximumConsumerBytes    = 128
	MaximumCorrelationBytes = 64
	MaximumKeyBytes         = 512
	MaximumTicketBytes      = 256
	MaximumReasonBytes      = 2_048
	MaximumReferenceBytes   = 2_048
	MaximumDocumentBytes    = 131_072
)

type EffectClass string

const (
	EffectDatabaseTransactional EffectClass = "database_transactional"
	EffectExternal              EffectClass = "external"
)

func (value EffectClass) valid() bool {
	return value == EffectDatabaseTransactional || value == EffectExternal
}
