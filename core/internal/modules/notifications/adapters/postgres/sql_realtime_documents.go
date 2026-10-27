package postgres

import _ "embed"

var (
	//go:embed sql/platform/insert_realtime_ticket.sql
	insertRealtimeTicketSQL string
	//go:embed sql/platform/consume_realtime_ticket.sql
	consumeRealtimeTicketSQL string
	//go:embed sql/platform/insert_legacy_connection_hit.sql
	insertLegacyConnectionHitSQL string
	//go:embed sql/platform/insert_provider_locator.sql
	insertProviderLocatorSQL string
)
