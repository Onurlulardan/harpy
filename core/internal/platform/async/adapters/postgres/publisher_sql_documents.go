package postgres

import _ "embed"

//go:embed sql/list_outbox_authorities.sql
var listOutboxAuthoritiesSQL string

//go:embed sql/claim_outbox.sql
var claimOutboxSQL string

//go:embed sql/heartbeat_publish.sql
var heartbeatPublishSQL string

//go:embed sql/confirm_publish.sql
var confirmPublishSQL string

//go:embed sql/release_publish.sql
var releasePublishSQL string

//go:embed sql/recover_expired_outbox.sql
var recoverExpiredOutboxSQL string

//go:embed sql/park_stale_outbox.sql
var parkStaleOutboxSQL string
