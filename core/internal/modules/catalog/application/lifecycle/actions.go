package lifecycle

type Action string

const (
	ActionRead       Action = "catalog.read"
	ActionInstall    Action = "catalog.install"
	ActionUpgrade    Action = "catalog.upgrade"
	ActionDisable    Action = "catalog.disable"
	ActionEnable     Action = "catalog.enable"
	ActionRestore    Action = "catalog.restore"
	ActionRetire     Action = "catalog.retire"
	ActionPurge      Action = "catalog.purge"
	ActionDemoManage Action = "catalog.demo.manage"
	ActionResume     Action = "catalog.operation.resume"
)

const (
	operationInstall          = "catalog.installation.install"
	operationUpgrade          = "catalog.installation.upgrade"
	operationRetry            = "catalog.operation.retry"
	operationCancel           = "catalog.operation.cancel"
	operationEnable           = "catalog.installation.enable"
	operationDisable          = "catalog.installation.disable"
	operationRestore          = "catalog.installation.restore"
	operationRetire           = "catalog.installation.retire"
	operationPurge            = "catalog.installation.purge"
	operationPurgeApprove     = "catalog.installation.purge.approve"
	operationGetOperation     = "catalog.operation.get"
	operationGetInstallation  = "catalog.installation.get"
	operationListInstallation = "catalog.installation.list"
	operationResume           = "catalog.operation.resume"
	operationDemoStatus       = "catalog.installation.demo.get"
	operationDemoPlan         = "catalog.installation.demo.removal_plan"
	operationDemoInstall      = "catalog.installation.demo.install"
	operationDemoRemove       = "catalog.installation.demo.remove"
)
