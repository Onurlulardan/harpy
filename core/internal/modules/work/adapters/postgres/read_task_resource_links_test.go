package postgres

import (
	"errors"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

type taskResourceRows struct {
	values [][]any
	index  int
	err    error
}

func (fixture *taskResourceRows) Next() bool {
	fixture.index++
	return fixture.index < len(fixture.values)
}
func (fixture *taskResourceRows) Err() error { return fixture.err }
func (*taskResourceRows) Close()             {}

func (fixture *taskResourceRows) Scan(targets ...any) error {
	for index, target := range targets {
		switch value := target.(type) {
		case *uuid.UUID:
			*value = fixture.values[fixture.index][index].(uuid.UUID)
		case *string:
			*value = fixture.values[fixture.index][index].(string)
		case *int64:
			*value = fixture.values[fixture.index][index].(int64)
		}
	}
	return nil
}

func TestTaskResourceLinkReadIsBoundedAndOwnerLocal(t *testing.T) {
	for _, fragment := range []string{
		"unnest($1::uuid[])", "work_resource_links", "work_resource_kind = 'TASK'",
		"state = 'ACTIVE'", "count(*) over", "link_rank <= 3",
	} {
		if !strings.Contains(strings.ToLower(readTaskResourceLinksSQL),
			strings.ToLower(fragment)) {
			t.Fatalf("resource link query missing %q", fragment)
		}
	}
	for _, forbidden := range []string{"forms_", "iam_", "workflow_", "for update"} {
		if strings.Contains(strings.ToLower(readTaskResourceLinksSQL), forbidden) {
			t.Fatalf("resource link query crossed owner boundary %q", forbidden)
		}
	}
}

func TestTaskResourceLinkScanPreservesExactTargetOrder(t *testing.T) {
	taskID, definitionID := testResourceID(1), testResourceID(2)
	versionID, recordID := testResourceID(3), testResourceID(4)
	rows := &taskResourceRows{index: -1, values: [][]any{
		resourceRow(testResourceID(5), taskID, "FORM", definitionID, versionID, 2),
		resourceRow(testResourceID(6), taskID, "FORM_RECORD", recordID, versionID, 2),
	}}
	result, err := scanTaskResourceLinks(rows, map[domain.ID]struct{}{taskID: {}})
	if err != nil || len(result[taskID]) != 2 ||
		result[taskID][0].TargetKind != "FORM" ||
		result[taskID][1].TargetKind != "FORM_RECORD" {
		t.Fatalf("exact targets drifted: result=%+v err=%v", result, err)
	}
}

func TestTaskResourceLinkScanRejectsExcessAndIncompleteRows(t *testing.T) {
	taskID, definitionID := testResourceID(1), testResourceID(2)
	versionID := testResourceID(3)
	for name, count := range map[string]int64{"excess": 3, "incomplete": 2} {
		t.Run(name, func(t *testing.T) {
			rows := &taskResourceRows{index: -1, values: [][]any{
				resourceRow(testResourceID(5), taskID, "FORM", definitionID, versionID, count),
			}}
			result, err := scanTaskResourceLinks(rows, map[domain.ID]struct{}{taskID: {}})
			if result != nil || !errors.Is(err, ports.ErrUnavailable) {
				t.Fatalf("corrupt cardinality escaped: result=%+v err=%v", result, err)
			}
		})
	}
}

func resourceRow(
	id, taskID domain.ID, kind string, targetID, versionID domain.ID, count int64,
) []any {
	return []any{uuid.UUID(id), uuid.UUID(taskID), "FORMS", kind,
		uuid.UUID(targetID), uuid.UUID(versionID), int64(1), int64(1), int64(1), count}
}

func testResourceID(seed byte) domain.ID {
	var value domain.ID
	value[0], value[6], value[8] = seed, 0x70, 0x80
	return value
}
