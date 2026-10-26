package link

import _ "embed"

//go:embed sql/tenant/register.sql
var registerSQL string

//go:embed sql/tenant/release.sql
var releaseSQL string

//go:embed sql/tenant/list.sql
var listSQL string

//go:embed sql/tenant/list_source.sql
var listSourceSQL string

//go:embed sql/tenant/lock_active_epoch.sql
var lockActiveEpochSQL string

//go:embed sql/tenant/load_form_record_adoption.sql
var loadFormRecordAdoptionSQL string

//go:embed sql/tenant/advance_form_record_adoption.sql
var advanceFormRecordAdoptionSQL string

//go:embed sql/tenant/seal_form_record_adoption.sql
var sealFormRecordAdoptionSQL string
