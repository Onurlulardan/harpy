package cataloghttp

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func RepresentPackage(value packageregistry.Package) Package {
	return Package{PackageID: value.ID, Slug: value.Slug, CreatedAt: value.CreatedAt}
}

func RepresentVersion(value packageregistry.SealedVersion) PackageVersion {
	return PackageVersion{
		PackageID: value.PackageID, Slug: value.Slug, Version: value.Version,
		ManifestSHA256: value.ManifestSHA256, KeyID: value.KeyID,
		RegisteredAt: value.RegisteredAt,
	}
}

func RepresentDependency(value packageregistry.DeclaredDependency) PackageDependency {
	return PackageDependency{
		PackageID: value.PackageID, VersionRange: value.VersionRange, Kind: value.Kind,
	}
}

func RepresentInstallation(value catalogdomain.Installation) Installation {
	var version *string
	if value.ActiveVersion.Valid() {
		wire := value.ActiveVersion.String()
		version = &wire
	}
	var lastOperation *string
	if value.LastOperationID.Valid() {
		wire := value.LastOperationID.String()
		lastOperation = &wire
	}
	return Installation{
		InstallationID: value.ID.String(), PackageID: value.PackageID.String(),
		ActiveVersion: version, State: string(value.State), ReferenceEpoch: value.ReferenceEpoch,
		Revision: value.RowVersion, LastOperationID: lastOperation,
	}
}

func RepresentOperation(value catalogdomain.Operation, replayed bool) Operation {
	return Operation{
		OperationID: value.ID.String(), Kind: string(value.Kind), State: string(value.State),
		PlanSHA256: value.Plan.SHA256, StartedAt: value.StartedAt,
		FailureCode: value.FailureCode, Revision: value.Revision, Replayed: replayed,
	}
}

func RepresentHandle(value lifecycleapp.OperationHandle) Operation {
	return RepresentOperation(value.Operation, value.Replayed)
}

func RepresentPurgeApproval(
	value lifecycleapp.PurgeApprovalReceipt,
	replayed bool,
) PurgeApproval {
	return PurgeApproval{
		ApprovalReceiptID: value.ID, OperationID: value.OperationID,
		RequestedBy: value.RequestedBy, ApprovedBy: value.ApprovedBy,
		ApprovalSHA256: value.ApprovalSHA256, ApprovedAt: value.ApprovedAt,
		ExpiresAt: value.ExpiresAt, Replayed: replayed,
	}
}

func RepresentDemoStatus(value lifecycleapp.DemoStatus) DemoStatus {
	components := make([]DemoComponentStatus, len(value.Components))
	for index, component := range value.Components {
		components[index] = DemoComponentStatus{
			ComponentID: component.ComponentID.String(),
			SourceRunID: component.SourceRunID.String(),
			RecordCount: len(component.RecordIDs),
		}
	}
	var updatedAt *time.Time
	if !value.UpdatedAt.IsZero() {
		updated := value.UpdatedAt
		updatedAt = &updated
	}
	return DemoStatus{InstallationID: value.InstallationID.String(),
		State: value.State, Components: components, UpdatedAt: updatedAt}
}

func RepresentDemoRemovalPlan(value lifecycleapp.DemoRemovalPlan) DemoRemovalPlan {
	components := make([]DemoRemovalComponent, len(value.Components))
	total := 0
	for index, component := range value.Components {
		ids := make([]string, len(component.RecordIDs))
		for recordIndex, recordID := range component.RecordIDs {
			ids[recordIndex] = recordID.String()
		}
		components[index] = DemoRemovalComponent{
			ComponentID: component.ComponentID.String(),
			SourceRunID: component.SourceRunID.String(), RecordIDs: ids,
		}
		total += len(ids)
	}
	return DemoRemovalPlan{InstallationID: value.InstallationID.String(),
		Components: components, RecordCount: total, PlanSHA256: value.PlanSHA256}
}

func RepresentMenu(value menuapp.Snapshot) Menu {
	items := make([]MenuItem, 0, len(value.Items))
	for _, item := range value.Items {
		var parent, target *string
		if item.ParentID != nil {
			wire := item.ParentID.String()
			parent = &wire
		}
		if item.TargetID != nil {
			wire := item.TargetID.String()
			target = &wire
		}
		items = append(items, MenuItem{
			ItemID: item.ID.String(), PackageID: item.PackageID.String(), ParentID: parent,
			Order: item.Order, Kind: string(item.Kind), Label: item.Label,
			TargetID: target, Visible: item.Visible,
		})
	}
	return Menu{
		VersionID: value.VersionID.String(), Number: value.Number,
		Revision: value.Revision, Items: items,
	}
}
