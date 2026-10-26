package item

const (
	OperationGet            = "documents.item.get"
	OperationList           = "documents.item.list"
	OperationListVersions   = "documents.item.versions.list"
	OperationUpdate         = "documents.item.update"
	OperationMove           = "documents.item.move"
	OperationDelete         = "documents.item.delete"
	OperationRestore        = "documents.item.restore"
	OperationPurge          = "documents.item.purge"
	OperationCancelPurge    = "documents.item.cancel_purge"
	OperationBatch          = "documents.item.batch"
	OperationLatest         = "documents.item.version.latest"
	OperationFolderContents = "documents.folder.contents"
)

const (
	PermissionRead       = "documents.item.read"
	PermissionUpdate     = "documents.item.update"
	PermissionMove       = "documents.item.move"
	PermissionDelete     = "documents.item.delete"
	PermissionPurge      = "documents.item.purge"
	PermissionFolderRead = "documents.folder.read"
)
