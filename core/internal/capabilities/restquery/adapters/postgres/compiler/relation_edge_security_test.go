package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type relationEdgeConsumer struct {
	statement RelationEdgeStatement
	err       error
}

func (consumer *relationEdgeConsumer) ConsumeRelationEdge(
	statement RelationEdgeStatement,
) error {
	consumer.statement = statement
	return consumer.err
}

func TestRelationEdgeBudgetsAndConsumerGate(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	if _, err := compiler.CompileRelationEdge(fixture.plan, graph, nodes[0], nil); !errors.Is(err, ErrContract) {
		t.Fatal("zero source identifiers were accepted")
	}
	oversized := make([]uuid.UUID, maximumIncludeSourceIDs+1)
	for index := range oversized {
		oversized[index] = uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprint(index)))
	}
	if _, err := compiler.CompileRelationEdge(
		fixture.plan, graph, nodes[0], oversized,
	); !errors.Is(err, ErrContract) {
		t.Fatal("oversized source identifier set was accepted")
	}
	statement, _ := compiler.CompileRelationEdge(
		fixture.plan, graph, nodes[0], includeSourceIDs,
	)
	consumer := &relationEdgeConsumer{}
	if err := ConsumeRelationEdgeStatement(
		statement, fixture.plan, graph, nodes[0], consumer,
	); err != nil || consumer.statement.Validate() != nil {
		t.Fatal("sealed relation statement was not consumed")
	}
	var nilConsumer *relationEdgeConsumer
	if !errors.Is(ConsumeRelationEdgeStatement(
		statement, fixture.plan, graph, nodes[0], nilConsumer,
	), ErrContract) {
		t.Fatal("typed nil relation consumer was accepted")
	}
}

func TestRelationEdgeTamperTransplantAndSerializationFailClosed(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	statement, _ := compiler.CompileRelationEdge(
		fixture.plan, graph, nodes[0], includeSourceIDs,
	)
	for name, mutate := range map[string]func(*RelationEdgeStatement){
		"sql":  func(value *RelationEdgeStatement) { value.sql += " OR TRUE" },
		"args": func(value *RelationEdgeStatement) { value.arguments[0].([]uuid.UUID)[0] = uuid.Nil },
		"seal": func(value *RelationEdgeStatement) { value.stateSeal[0] ^= 0xff },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := statement
			tampered.arguments = cloneArguments(statement.arguments)
			mutate(&tampered)
			if !errors.Is(tampered.Validate(), ErrContract) {
				t.Fatal("tampered relation statement was accepted")
			}
		})
	}
	if !errors.Is(statement.ValidatePlan(fixture.plan, graph, nodes[1]), ErrContract) {
		t.Fatal("relation statement node transplant was accepted")
	}
	other := newIncludePlanFixtureForTenant(
		t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab"),
	)
	if !errors.Is(statement.ValidatePlan(other.plan, graph, nodes[0]), ErrContract) {
		t.Fatal("relation statement plan transplant was accepted")
	}
	if fmt.Sprint(statement) != compilerRedaction {
		t.Fatal("relation statement formatting leaked")
	}
	if document, err := json.Marshal(statement); err == nil || document != nil {
		t.Fatal("relation statement JSON did not fail closed")
	}
}
