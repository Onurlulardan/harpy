package worker

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type notificationsDatabaseJobStoreStub struct{ jobexecution.DatabaseJobStore }

func TestNotificationsRetentionProcessorIsComposed(t *testing.T) {
	release, err := platformasync.LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	store := &notificationsDatabaseJobStoreStub{}
	processors, err := buildNotificationsRetentionProcessors(release, store)
	if err != nil {
		t.Fatal(err)
	}
	if len(processors) != 1 || processors[0] == nil {
		t.Fatalf("Notifications retention processors = %#v", processors)
	}
}
