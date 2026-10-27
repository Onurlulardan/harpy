package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractRecordRequirements(
	collector *requirementCollector,
	catalogs api.CatalogSet,
	root api.FieldCatalog,
	rootPath resolved.RequestPath,
	record resolved.RecordQuery,
) error {
	projections, projectionErr := record.Projections()
	sorts, sortErr := record.Sorts()
	search, searchErr := record.Search()
	if projectionErr != nil || sortErr != nil || searchErr != nil ||
		extractProjections(collector, root, rootPath, projections) != nil ||
		extractSorts(collector, root, rootPath, sorts) != nil ||
		extractSearch(collector, root, rootPath, search) != nil {
		return errPolicyEvaluationContract
	}
	where, hasWhere, err := record.Where()
	if err != nil || hasWhere && extractFilter(
		collector, root, rootPath, where,
	) != nil {
		return errPolicyEvaluationContract
	}
	count, countErr := record.Count()
	if countErr != nil {
		return errPolicyEvaluationContract
	}
	if count == api.CountModeExact {
		path, pathErr := record.CountPath()
		if pathErr != nil || collector.addAction(
			RequirementExactCount, root, rootPath, path,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	includes, err := record.Includes()
	if err != nil {
		return errPolicyEvaluationContract
	}
	for _, include := range includes {
		if err := extractIncludeRequirements(
			collector, catalogs, root, rootPath, include,
		); err != nil {
			return err
		}
	}
	return nil
}
