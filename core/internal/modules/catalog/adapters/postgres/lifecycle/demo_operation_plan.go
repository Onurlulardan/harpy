package lifecycle

import (
	"context"
	_ "embed"
	"sort"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/insert_demo_operation_plan.sql
var insertDemoOperationPlanSQL string

//go:embed sql/tenant/read_demo_operation_plan.sql
var readDemoOperationPlanSQL string

func (storage *store) RecordDemoOperationPlan(
	ctx context.Context,
	plan catalogapp.DemoOperationPlan,
) (catalogapp.DemoOperationPlan, error) {
	plan = canonicalDemoOperationPlan(plan)
	if !validDemoOperationPlan(plan) {
		return catalogapp.DemoOperationPlan{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return catalogapp.DemoOperationPlan{}, unavailable()
	}
	for _, component := range plan.Components {
		recordIDs, err := encodeDemoRecordIDs(component.RecordIDs)
		if err != nil {
			return catalogapp.DemoOperationPlan{}, err
		}
		if _, err = storage.execute(ctx, insertDemoOperationPlanSQL,
			plan.OperationID, plan.InstallationID, plan.Action,
			component.ComponentID, component.SourceRunID, recordIDs,
			plan.ExpectedPlanSHA256, plan.CreatedAt,
		); err != nil {
			return catalogapp.DemoOperationPlan{}, unavailable()
		}
	}
	stored, err := storage.ReadDemoOperationPlan(ctx, plan.OperationID)
	if err != nil || !sameDemoOperationPlan(stored, plan) {
		return catalogapp.DemoOperationPlan{}, catalogapp.ErrConflict
	}
	return stored, nil
}

func (storage *store) ReadDemoOperationPlan(
	ctx context.Context,
	operationID uuid.UUID,
) (catalogapp.DemoOperationPlan, error) {
	if operationID.Version() != 7 {
		return catalogapp.DemoOperationPlan{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return catalogapp.DemoOperationPlan{}, unavailable()
	}
	rows, err := storage.query(ctx, readDemoOperationPlanSQL, operationID)
	if err != nil {
		return catalogapp.DemoOperationPlan{}, unavailable()
	}
	defer rows.Close()
	var result catalogapp.DemoOperationPlan
	for rows.Next() {
		var rowOperation, installationID, componentID, sourceRunID uuid.UUID
		var action, expected string
		var recordIDs []byte
		var createdAt time.Time
		if err = rows.Scan(&rowOperation, &installationID, &action, &componentID,
			&sourceRunID, &recordIDs, &expected, &createdAt); err != nil {
			return catalogapp.DemoOperationPlan{}, unavailable()
		}
		ids, decodeErr := decodeDemoRecordIDs(recordIDs)
		if decodeErr != nil {
			return catalogapp.DemoOperationPlan{}, unavailable()
		}
		if result.OperationID == uuid.Nil {
			result = catalogapp.DemoOperationPlan{OperationID: rowOperation,
				InstallationID: installationID, Action: action,
				ExpectedPlanSHA256: expected, CreatedAt: createdAt.UTC()}
		} else if result.OperationID != rowOperation ||
			result.InstallationID != installationID || result.Action != action ||
			result.ExpectedPlanSHA256 != expected || !result.CreatedAt.Equal(createdAt.UTC()) {
			return catalogapp.DemoOperationPlan{}, unavailable()
		}
		result.Components = append(result.Components, catalogapp.DemoRemovalComponent{
			ComponentID: componentID, SourceRunID: sourceRunID, RecordIDs: ids,
		})
	}
	if rows.Err() != nil || result.OperationID == uuid.Nil || !validDemoOperationPlan(result) {
		return catalogapp.DemoOperationPlan{}, catalogapp.ErrNotFound
	}
	return result, nil
}

func canonicalDemoOperationPlan(plan catalogapp.DemoOperationPlan) catalogapp.DemoOperationPlan {
	plan.CreatedAt = plan.CreatedAt.UTC()
	plan.Components = append([]catalogapp.DemoRemovalComponent(nil), plan.Components...)
	for index := range plan.Components {
		ids, _ := canonicalDemoRecordIDs(plan.Components[index].RecordIDs)
		plan.Components[index].RecordIDs = ids
	}
	sort.Slice(plan.Components, func(i, j int) bool {
		return plan.Components[i].ComponentID.String() < plan.Components[j].ComponentID.String()
	})
	return plan
}

func sameDemoOperationPlan(left, right catalogapp.DemoOperationPlan) bool {
	if left.OperationID != right.OperationID || left.InstallationID != right.InstallationID ||
		left.Action != right.Action || left.ExpectedPlanSHA256 != right.ExpectedPlanSHA256 ||
		!left.CreatedAt.Equal(right.CreatedAt) || len(left.Components) != len(right.Components) {
		return false
	}
	for index := range left.Components {
		if left.Components[index].ComponentID != right.Components[index].ComponentID ||
			left.Components[index].SourceRunID != right.Components[index].SourceRunID ||
			!sameDemoUUIDs(left.Components[index].RecordIDs, right.Components[index].RecordIDs) {
			return false
		}
	}
	return true
}
