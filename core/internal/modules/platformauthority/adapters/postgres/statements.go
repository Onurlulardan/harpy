package postgres

import _ "embed"

var (
	//go:embed sql/platform/bind.sql
	bindSQL string
	//go:embed sql/platform/get_binding.sql
	getBindingSQL string
	//go:embed sql/platform/get_saas_binding.sql
	getSaaSBindingSQL string
	//go:embed sql/platform/prepare.sql
	prepareSQL string
	//go:embed sql/platform/record_ack.sql
	recordAckSQL string
	//go:embed sql/platform/commit.sql
	commitSQL string
	//go:embed sql/platform/get_operation.sql
	getOperationSQL string
	//go:embed sql/platform/get_head.sql
	getHeadSQL string
	//go:embed sql/platform/activate_signer.sql
	activateSignerSQL string
	//go:embed sql/platform/fence_signer.sql
	fenceSignerSQL string
	//go:embed sql/platform/installation_bind_owner.sql
	installationBindOwnerSQL string
	//go:embed sql/platform/installation_pin_head.sql
	installationPinHeadSQL string
	//go:embed sql/platform/development_binding_input.sql
	developmentBindingInputSQL string
	//go:embed sql/platform/claim_delivery.sql
	claimDeliverySQL string
	//go:embed sql/platform/retry_delivery.sql
	retryDeliverySQL string
	//go:embed sql/platform/await_delivery_import.sql
	awaitDeliveryImportSQL string
	//go:embed sql/platform/succeed_delivery.sql
	succeedDeliverySQL string
)
