package postgres

import (
	"strings"
	"testing"
)

func TestConsumerSQLDocumentsArePresentAndOperationSeparated(t *testing.T) {
	t.Parallel()
	documents := map[string]string{
		"lock consumer stream":     lockConsumerStreamSQL,
		"assert consumer producer": assertConsumerProducerSQL,
		"inspect event":            inspectEventSQL, "insert event receipt": insertEventReceiptSQL,
		"insert event watermark":     insertEventWatermarkSQL,
		"verify external receipt":    verifyExternalEffectReceiptSQL,
		"inspect bootstrap":          inspectBootstrapSQL,
		"insert bootstrap watermark": insertBootstrapWatermarkSQL,
	}
	for name, document := range documents {
		if strings.TrimSpace(document) == "" {
			t.Fatalf("%s SQL is empty", name)
		}
	}
	if !strings.Contains(lockConsumerStreamSQL, "pg_advisory_xact_lock") ||
		!strings.Contains(lockConsumerStreamSQL, "platform_async_assert_current_authority_v1") ||
		!strings.Contains(inspectEventSQL, "producer_missing") ||
		!strings.Contains(inspectBootstrapSQL, "stream_not_empty") {
		t.Fatal("consumer pre-effect fence is incomplete")
	}
	if !strings.Contains(insertEventReceiptSQL, "'event_receipt'") ||
		!strings.Contains(insertEventWatermarkSQL, "'event'") {
		t.Fatal("event receipt and watermark are not separate writes")
	}
}
