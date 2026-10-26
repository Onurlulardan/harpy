package iam

import (
	"sync/atomic"

	maintenancepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/maintenance"
	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type MaintenanceJobKey struct {
	Kind            string
	PayloadRevision int
}

type MaintenanceJobRuntime struct {
	adapter *maintenancepostgres.Runtime
	binding platformasync.DatabaseJobBinding
	closed  atomic.Bool
}

func NewMaintenanceJobRuntime(
	keys []MaintenanceDirectoryKey,
) (*MaintenanceJobRuntime, error) {
	// Ownership is transferred to the runtime. The caller's backing array is
	// consumed and wiped on every return path.
	defer wipeMaintenanceDirectoryKeyInputs(keys)
	adapterKeys := make([]maintenancepostgres.DirectoryKey, len(keys))
	for index, key := range keys {
		adapterKeys[index] = maintenancepostgres.DirectoryKey{
			ID: key.id, Material: key.material, Active: key.active,
		}
	}
	adapter, err := maintenancepostgres.NewRuntime(adapterKeys)
	clearMaintenanceDirectoryKeys(adapterKeys)
	if err != nil {
		return nil, err
	}
	binding, err := platformasync.NewDatabaseJobBinding(
		"iam",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store := maintenancepostgres.NewStore(adapter,
				maintenancepostgres.DatabaseAccess{
					Query: database.Query, QueryRow: database.QueryRow,
				})
			commands, commandErr := maintenanceapp.NewCommands(store, uuid.NewV7)
			return &maintenanceDatabaseJobCommand{
				commands: commands, buildErr: commandErr,
			}
		},
	)
	if err != nil {
		_ = adapter.Close()
		return nil, err
	}
	return &MaintenanceJobRuntime{adapter: adapter, binding: binding}, nil
}

func wipeMaintenanceDirectoryKeyInputs(keys []MaintenanceDirectoryKey) {
	for index := range keys {
		keys[index].id = ""
		clear(keys[index].material[:])
		keys[index].active = false
	}
}

func (runtime *MaintenanceJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil || runtime.closed.Load() {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}

func (runtime *MaintenanceJobRuntime) JobKeys() []MaintenanceJobKey {
	return MaintenanceJobKeys()
}

func (runtime *MaintenanceJobRuntime) Close() error {
	if runtime == nil || !runtime.closed.CompareAndSwap(false, true) {
		return nil
	}
	return runtime.adapter.Close()
}

func clearMaintenanceDirectoryKeys(keys []maintenancepostgres.DirectoryKey) {
	for index := range keys {
		clear(keys[index].Material[:])
	}
}
