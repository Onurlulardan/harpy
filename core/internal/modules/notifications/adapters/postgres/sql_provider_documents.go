package postgres

import _ "embed"

var (
	//go:embed sql/tenant/resolve_provider.sql
	resolveProviderSQL string
	//go:embed sql/tenant/get_provider_config.sql
	getProviderConfigSQL string
	//go:embed sql/tenant/insert_provider_secret.sql
	insertProviderSecretSQL string
	//go:embed sql/tenant/get_provider_secret.sql
	getProviderSecretSQL string
	//go:embed sql/tenant/insert_provider_config.sql
	insertProviderConfigSQL string
	//go:embed sql/tenant/update_provider_config.sql
	updateProviderConfigSQL string
	//go:embed sql/tenant/rotate_provider_secret.sql
	rotateProviderSecretSQL string
	//go:embed sql/tenant/rewrap_provider_secret.sql
	rewrapProviderSecretSQL string
)
