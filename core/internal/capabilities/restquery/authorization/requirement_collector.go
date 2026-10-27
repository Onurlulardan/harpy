package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type requirementCollector struct {
	requirements []Requirement
	resources    map[[sha256.Size]byte]resourceRequirement
	identities   map[[sha256.Size]byte]struct{}
}

func newRequirementCollector() *requirementCollector {
	return &requirementCollector{
		requirements: make([]Requirement, 0),
		resources:    make(map[[sha256.Size]byte]resourceRequirement),
		identities:   make(map[[sha256.Size]byte]struct{}),
	}
}

func (collector *requirementCollector) addResource(
	catalog api.FieldCatalog,
	path resolved.RequestPath,
) error {
	requirement, err := collector.add(
		RequirementResourceRead, catalog, path, path, uuid.Nil, uuid.Nil,
	)
	if err != nil {
		return err
	}
	if _, duplicate := collector.resources[requirement.id]; duplicate {
		return errPolicyEvaluationContract
	}
	collector.resources[requirement.id] = resourceRequirement{
		requirement: requirement, catalog: catalog, path: path,
	}
	return nil
}

func (collector *requirementCollector) addField(
	kind RequirementKind,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
	field api.FieldDefinition,
) error {
	identifier, err := field.ID()
	if err != nil {
		return errPolicyEvaluationContract
	}
	_, err = collector.add(
		kind, catalog, resourcePath, requestPath, identifier, uuid.Nil,
	)
	return err
}

func (collector *requirementCollector) addAction(
	kind RequirementKind,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
) error {
	_, err := collector.add(
		kind, catalog, resourcePath, requestPath, uuid.Nil, uuid.Nil,
	)
	return err
}

func (collector *requirementCollector) addRelation(
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
	relation api.RelationDefinition,
) error {
	identifier, err := relation.RelationID()
	if err != nil {
		return errPolicyEvaluationContract
	}
	_, err = collector.add(
		RequirementRelationInclude, catalog, resourcePath, requestPath,
		uuid.Nil, identifier,
	)
	return err
}
