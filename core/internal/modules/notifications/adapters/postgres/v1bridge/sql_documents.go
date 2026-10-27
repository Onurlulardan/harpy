package v1bridge

import _ "embed"

var (
	//go:embed sql/tenant/lock_domain_authority.sql
	lockDomainAuthoritySQL string
	//go:embed sql/tenant/lock_writer_authority.sql
	lockWriterAuthoritySQL string
	//go:embed sql/tenant/find_intent.sql
	findIntentSQL string
	//go:embed sql/tenant/insert_intent.sql
	insertIntentSQL string
)
