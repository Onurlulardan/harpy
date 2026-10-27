package notificationaccess

import _ "embed"

//go:embed sql/tenant/read_session_selector.sql
var readSessionSelectorSQL string
