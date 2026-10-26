package work

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/boards"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/calendar"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/checklists"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/columns"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/comments"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/labels"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/memberships"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/recurrence"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func buildHTTPServices(mutator ports.Mutator,
	principals ports.PrincipalResolver,
	resources ports.ResourceAuthority,
	references ports.ResourceReferenceRegistry,
) (result httpServices, err error) {
	commonClock, commonIDs := runtimeClock{}, runtimeIDs{}
	if result.boards, err = boards.New(boards.Dependencies{
		Mutator: mutator, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.columns, err = columns.New(columns.Dependencies{
		Mutator: mutator, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.memberships, err = memberships.New(memberships.Dependencies{
		Mutator: mutator, Principals: principals,
		Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.tasks, err = tasks.New(tasks.Dependencies{Mutator: mutator,
		Principals: principals, Resources: resources,
		References: references,
		Clock:      commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.comments, err = comments.New(comments.Dependencies{
		Mutator: mutator, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.labels, err = labels.New(labels.Dependencies{
		Mutator: mutator, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.checklists, err = checklists.New(checklists.Dependencies{
		Mutator: mutator, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	if result.calendar, err = calendar.New(calendar.Dependencies{Mutator: mutator,
		Principals: principals, Clock: commonClock, IDs: commonIDs}); err != nil {
		return result, err
	}
	result.recurrence, err = recurrence.New(recurrence.Dependencies{Mutator: mutator,
		TZDB: runtimeTZDB{}, Clock: commonClock, IDs: commonIDs})
	return result, err
}
