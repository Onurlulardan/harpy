package runtimev1

import (
	"encoding/json"
	"fmt"
)

type ReadinessStatus string

const (
	ReadinessStatusNotReady ReadinessStatus = "not_ready"
	ReadinessStatusReady    ReadinessStatus = "ready"
)

type ReadinessReason string

const (
	ReadinessReasonDependencyFailed ReadinessReason = "dependency_failed"
	ReadinessReasonDraining         ReadinessReason = "draining"
	ReadinessReasonNone             ReadinessReason = "none"
	ReadinessReasonStarting         ReadinessReason = "starting"
)

type ReadinessCheckName string

const (
	ReadinessCheckDatabase         ReadinessCheckName = "database"
	ReadinessCheckDatabaseRole     ReadinessCheckName = "database_role"
	ReadinessCheckDeploymentPolicy ReadinessCheckName = "deployment_policy"
	ReadinessCheckSchemaContract   ReadinessCheckName = "schema_contract"
)

type ReadinessCheck struct {
	Name   ReadinessCheckName `json:"name"`
	Status ReadinessStatus    `json:"status"`
}

type ReadinessStates struct {
	Database         ReadinessStatus
	DatabaseRole     ReadinessStatus
	DeploymentPolicy ReadinessStatus
	SchemaContract   ReadinessStatus
}

type Readyz struct {
	Status  ReadinessStatus   `json:"status"`
	Service Service           `json:"service"`
	Reason  ReadinessReason   `json:"reason"`
	Checks  [4]ReadinessCheck `json:"checks"`
}

func NewReadyz(
	service Service,
	reason ReadinessReason,
	states ReadinessStates,
) (Readyz, error) {
	readyz := Readyz{
		Status:  ReadinessStatusNotReady,
		Service: service,
		Reason:  reason,
		Checks: [4]ReadinessCheck{
			{Name: ReadinessCheckDatabase, Status: states.Database},
			{Name: ReadinessCheckDatabaseRole, Status: states.DatabaseRole},
			{Name: ReadinessCheckDeploymentPolicy, Status: states.DeploymentPolicy},
			{Name: ReadinessCheckSchemaContract, Status: states.SchemaContract},
		},
	}
	if reason == ReadinessReasonNone && readyz.allChecksReady() {
		readyz.Status = ReadinessStatusReady
	}
	if err := readyz.Validate(); err != nil {
		return Readyz{}, err
	}
	return readyz, nil
}

func (readyz Readyz) Validate() error {
	if !validReadinessStatus(readyz.Status) {
		return fmt.Errorf("readiness status is invalid")
	}
	if !readyz.Service.Valid() {
		return fmt.Errorf("readiness service is invalid")
	}
	if !validReadinessReason(readyz.Reason) {
		return fmt.Errorf("readiness reason is invalid")
	}
	expectedNames := [4]ReadinessCheckName{
		ReadinessCheckDatabase,
		ReadinessCheckDatabaseRole,
		ReadinessCheckDeploymentPolicy,
		ReadinessCheckSchemaContract,
	}
	for index, check := range readyz.Checks {
		if check.Name != expectedNames[index] {
			return fmt.Errorf("readiness check %d name or order is invalid", index)
		}
		if !validReadinessStatus(check.Status) {
			return fmt.Errorf("readiness check %d status is invalid", index)
		}
	}
	allReady := readyz.allChecksReady()
	if readyz.Status == ReadinessStatusReady {
		if readyz.Reason != ReadinessReasonNone || !allReady {
			return fmt.Errorf("ready status requires reason none and ready checks")
		}
		return nil
	}
	if readyz.Reason == ReadinessReasonNone {
		return fmt.Errorf("not-ready status requires a non-none reason")
	}
	if readyz.Reason == ReadinessReasonDependencyFailed && allReady {
		return fmt.Errorf("dependency failure requires a failed check")
	}
	return nil
}

func (readyz Readyz) MarshalJSON() ([]byte, error) {
	if err := readyz.Validate(); err != nil {
		return nil, err
	}
	type wireReadyz Readyz
	return json.Marshal(wireReadyz(readyz))
}

func (readyz Readyz) allChecksReady() bool {
	for _, check := range readyz.Checks {
		if check.Status != ReadinessStatusReady {
			return false
		}
	}
	return true
}

func validReadinessStatus(status ReadinessStatus) bool {
	return status == ReadinessStatusReady || status == ReadinessStatusNotReady
}

func validReadinessReason(reason ReadinessReason) bool {
	switch reason {
	case ReadinessReasonDependencyFailed, ReadinessReasonDraining,
		ReadinessReasonNone, ReadinessReasonStarting:
		return true
	default:
		return false
	}
}
