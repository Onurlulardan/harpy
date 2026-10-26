package policy

import (
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func requirementAction(
	requirement restauthorization.Requirement,
	published catalog.Published,
) (string, error) {
	kind, err := requirement.Kind()
	if err != nil {
		return "", catalog.ErrContract
	}
	switch kind {
	case restauthorization.RequirementResourceRead,
		restauthorization.RequirementRelationInclude:
		return "read", nil
	case restauthorization.RequirementExactCount:
		return "query.count.exact", nil
	}
	fieldID, present, err := requirement.FieldID()
	if err != nil || !present {
		return "", catalog.ErrContract
	}
	classification, err := published.Classification(fieldID)
	if err != nil {
		return "", catalog.ErrContract
	}
	return classificationAction(classification)
}

func classificationAction(classification profile.Classification) (string, error) {
	switch classification {
	case profile.ClassificationPublic, profile.ClassificationInternal:
		return "read", nil
	case profile.ClassificationConfidential:
		return "read.confidential", nil
	case profile.ClassificationRestricted:
		return "read.restricted", nil
	default:
		return "", catalog.ErrContract
	}
}
