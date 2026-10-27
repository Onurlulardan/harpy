package systemtransfer

import _ "embed"

var (
	//go:embed sql/tenant/find_import_user.sql
	findImportUserSQL string
	//go:embed sql/tenant/find_import_role.sql
	findImportRoleSQL string
	//go:embed sql/tenant/find_import_organization.sql
	findImportOrganizationSQL string
	//go:embed sql/tenant/claim_import_worker.sql
	claimImportWorkerSQL string
	//go:embed sql/tenant/read_import_worker_receipts.sql
	readImportWorkerReceiptsSQL string
	//go:embed sql/tenant/commit_import_worker_batch.sql
	commitImportWorkerBatchSQL string
	//go:embed sql/tenant/finish_import_worker.sql
	finishImportWorkerSQL string
	//go:embed sql/tenant/resolve_import_role.sql
	resolveImportRoleSQL string
)
