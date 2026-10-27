package postgres

import _ "embed"

//go:embed sql/lock_consumer_stream.sql
var lockConsumerStreamSQL string

//go:embed sql/assert_consumer_producer.sql
var assertConsumerProducerSQL string

//go:embed sql/inspect_event.sql
var inspectEventSQL string

//go:embed sql/park_event_gap.sql
var parkEventGapSQL string

//go:embed sql/insert_event_receipt.sql
var insertEventReceiptSQL string

//go:embed sql/insert_event_watermark.sql
var insertEventWatermarkSQL string

//go:embed sql/verify_external_effect_receipt.sql
var verifyExternalEffectReceiptSQL string

//go:embed sql/inspect_bootstrap.sql
var inspectBootstrapSQL string

//go:embed sql/insert_bootstrap_watermark.sql
var insertBootstrapWatermarkSQL string
