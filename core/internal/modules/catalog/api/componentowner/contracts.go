package componentowner

import contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
import lifecyclecontract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"

type Component = contract.Component
type Readiness = contract.Readiness
type ProvisionRequest = contract.ProvisionRequest
type ProvisionReceipt = contract.ProvisionReceipt
type Command = contract.Command
type Receipt = contract.Receipt
type Blocker = contract.Blocker
type BlockerReceipt = contract.BlockerReceipt
type LifecycleCommand = lifecyclecontract.LifecycleCommand
type PurgeAuthority = lifecyclecontract.PurgeAuthority
type PurgeCommand = lifecyclecontract.PurgeCommand
type OwnerResult = lifecyclecontract.OwnerResult

const (
	ProvisionAvailable = contract.ProvisionAvailable
	ProvisionStaged    = contract.ProvisionStaged
	ProvisionOperation = "catalog.operation.resume"
)
