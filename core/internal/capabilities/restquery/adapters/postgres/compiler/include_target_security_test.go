package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type includeTargetConsumer struct {
	statement IncludeTargetStatement
	err       error
}

func (consumer *includeTargetConsumer) ConsumeIncludeTarget(
	statement IncludeTargetStatement,
) error {
	consumer.statement = statement
	return consumer.err
}

func TestIncludeTargetBudgetsAndConsumerGate(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	if _, err := compiler.CompileIncludeTarget(fixture.plan, graph, nodes[0], nil); !errors.Is(err, ErrContract) {
		t.Fatal("zero target candidates were accepted")
	}
	oversized := make([]uuid.UUID, maximumIncludeCandidateIDs+1)
	for index := range oversized {
		oversized[index] = includeCandidateIDs[0]
	}
	if _, err := compiler.CompileIncludeTarget(
		fixture.plan, graph, nodes[0], oversized,
	); !errors.Is(err, ErrContract) {
		t.Fatal("oversized target candidate set was accepted")
	}
	statement, _ := compiler.CompileIncludeTarget(
		fixture.plan, graph, nodes[0], includeCandidateIDs,
	)
	consumer := &includeTargetConsumer{}
	if err := ConsumeIncludeTargetStatement(
		statement, fixture.plan, graph, nodes[0], consumer,
	); err != nil || consumer.statement.Validate() != nil {
		t.Fatal("sealed include-target statement was not consumed")
	}
	var nilConsumer *includeTargetConsumer
	if !errors.Is(ConsumeIncludeTargetStatement(
		statement, fixture.plan, graph, nodes[0], nilConsumer,
	), ErrContract) {
		t.Fatal("typed nil include-target consumer was accepted")
	}
}

func TestIncludeTargetTamperTransplantAndSerializationFailClosed(t *testing.T) {
	fixture := newIncludePlanFixture(t)
	compiler := includeCompiler(t)
	graph, _ := compiler.CompileIncludeGraph(fixture.plan)
	nodes, _ := graph.Nodes()
	statement, _ := compiler.CompileIncludeTarget(
		fixture.plan, graph, nodes[0], includeCandidateIDs,
	)
	for name, mutate := range map[string]func(*IncludeTargetStatement){
		"sql":    func(value *IncludeTargetStatement) { value.sql += " LIMIT 1" },
		"args":   func(value *IncludeTargetStatement) { value.arguments[0].([]uuid.UUID)[0] = uuid.Nil },
		"column": func(value *IncludeTargetStatement) { value.columns[0].outputKey = "changed" },
		"index":  func(value *IncludeTargetStatement) { value.stableIndex = 1 },
		"seal":   func(value *IncludeTargetStatement) { value.stateSeal[0] ^= 0xff },
	} {
		t.Run(name, func(t *testing.T) {
			tampered := statement
			tampered.arguments = cloneArguments(statement.arguments)
			tampered.columns = append([]ResultColumn{}, statement.columns...)
			mutate(&tampered)
			if !errors.Is(tampered.Validate(), ErrContract) {
				t.Fatal("tampered include-target statement was accepted")
			}
		})
	}
	if !errors.Is(statement.ValidatePlan(fixture.plan, graph, nodes[1]), ErrContract) {
		t.Fatal("include-target node transplant was accepted")
	}
	other := newIncludePlanFixtureForTenant(
		t, uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab"),
	)
	if !errors.Is(statement.ValidatePlan(other.plan, graph, nodes[0]), ErrContract) {
		t.Fatal("include-target plan transplant was accepted")
	}
	if fmt.Sprint(statement) != compilerRedaction {
		t.Fatal("include-target formatting leaked")
	}
	if document, err := json.Marshal(statement); err == nil || document != nil {
		t.Fatal("include-target JSON did not fail closed")
	}
}
