package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractIncludeRequirements(
	collector *requirementCollector,
	catalogs api.CatalogSet,
	source api.FieldCatalog,
	sourcePath resolved.RequestPath,
	include resolved.Include,
) error {
	path, pathErr := include.Path()
	relation, relationErr := include.Relation()
	relationPath, memberErr := requestMemberPath(path, "/relation")
	targetReference, targetErr := include.Target()
	if pathErr != nil || relationErr != nil || memberErr != nil || targetErr != nil ||
		collector.addRelation(
			source, sourcePath, relationPath, relation,
		) != nil {
		return errPolicyEvaluationContract
	}
	target, found, err := catalogs.Catalog(targetReference)
	if err != nil || !found || collector.addResource(target, path) != nil {
		return errPolicyEvaluationContract
	}
	projections, projectionErr := include.Projections()
	sorts, sortErr := include.Sorts()
	if projectionErr != nil || sortErr != nil ||
		extractProjections(collector, target, path, projections) != nil ||
		extractSorts(collector, target, path, sorts) != nil {
		return errPolicyEvaluationContract
	}
	where, hasWhere, whereErr := include.Where()
	if whereErr != nil || hasWhere && extractFilter(
		collector, target, path, where,
	) != nil {
		return errPolicyEvaluationContract
	}
	children, err := include.Children()
	if err != nil {
		return errPolicyEvaluationContract
	}
	for _, child := range children {
		if err := extractIncludeRequirements(
			collector, catalogs, target, path, child,
		); err != nil {
			return err
		}
	}
	return nil
}
