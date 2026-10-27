package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type includeStatementRole uint8

const (
	includeStatementRoleRelationEdge includeStatementRole = iota + 1
	includeStatementRoleTargetRecord
)

type includeStatementBinding struct {
	plan      includePlanBinding
	graphSeal [sha256.Size]byte
	nodeSeal  [sha256.Size]byte
	role      includeStatementRole
	validated bool
}

func newIncludeStatementBinding(
	graph IncludeGraph, node IncludeNode, role includeStatementRole,
) (includeStatementBinding, error) {
	if graph.Validate() != nil || validateGraphNode(graph, node) != nil ||
		role < includeStatementRoleRelationEdge || role > includeStatementRoleTargetRecord {
		return includeStatementBinding{}, ErrContract
	}
	binding := includeStatementBinding{
		plan: graph.binding, graphSeal: graph.stateSeal, nodeSeal: node.stateSeal,
		role: role, validated: true,
	}
	if binding.validate() != nil {
		return includeStatementBinding{}, ErrContract
	}
	return binding, nil
}

func (binding includeStatementBinding) validate() error {
	if !binding.validated || binding.plan.validate() != nil ||
		binding.graphSeal == [sha256.Size]byte{} || binding.nodeSeal == [sha256.Size]byte{} ||
		binding.role < includeStatementRoleRelationEdge ||
		binding.role > includeStatementRoleTargetRecord {
		return ErrContract
	}
	return nil
}

func (binding includeStatementBinding) validatePlan(
	plan resolved.AuthorizedPlan, graph IncludeGraph, node IncludeNode,
) error {
	expectedPlan, err := newIncludePlanBinding(plan)
	expected, expectedErr := newIncludeStatementBinding(graph, node, binding.role)
	if err != nil || expectedErr != nil || graph.ValidatePlan(plan) != nil ||
		expectedPlan != binding.plan || expected != binding {
		return ErrContract
	}
	return nil
}
