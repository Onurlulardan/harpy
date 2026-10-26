package iam

import (
	"fmt"

	notificationaccess "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/notificationaccess"
	notificationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/notificationaccess"
)

type NotificationsAccessRuntime struct {
	shared   *FormsRecordAccessRuntime
	sessions notificationapp.SessionReader
}

func NewNotificationsAccessRuntime(shared *FormsRecordAccessRuntime) (
	*NotificationsAccessRuntime, error,
) {
	if shared == nil || shared.pool == nil || shared.coordinator == nil ||
		shared.security == nil || shared.clock == nil {
		return nil, fmt.Errorf("IAM Notifications access dependencies are required")
	}
	sessions, err := notificationaccess.NewReader()
	if err != nil {
		return nil, fmt.Errorf("build IAM Notifications session reader: %w", err)
	}
	return &NotificationsAccessRuntime{shared: shared, sessions: sessions}, nil
}
