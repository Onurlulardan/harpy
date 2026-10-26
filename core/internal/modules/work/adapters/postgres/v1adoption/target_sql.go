package v1adoption

import _ "embed"

//go:embed sql/tenant/begin_adoption.sql
var beginAdoptionSQL string

//go:embed sql/tenant/fence_v1_writer.sql
var fenceWriterSQL string

//go:embed sql/tenant/set_adoption_state.sql
var setAdoptionStateSQL string

//go:embed sql/tenant/record_adoption_object.sql
var recordObjectSQL string

//go:embed sql/tenant/record_adoption_issue.sql
var recordIssueSQL string

//go:embed sql/tenant/list_adoption_objects.sql
var listObjectsSQL string

//go:embed sql/tenant/accept_adoption.sql
var acceptAdoptionSQL string

//go:embed sql/tenant/read_adoption_evidence.sql
var readEvidenceSQL string

//go:embed sql/tenant/accept_reverse.sql
var acceptReverseSQL string

//go:embed sql/tenant/reopen_v1_writer.sql
var reopenWriterSQL string

//go:embed sql/tenant/retire_v1_writer.sql
var retireWriterSQL string
