package usage

import _ "embed"

var (
	//go:embed sql/platform/resolve_authority.sql
	resolveAuthoritySQL string
	//go:embed sql/platform/assert_authority.sql
	assertAuthoritySQL string
	//go:embed sql/platform/resolve_period.sql
	resolvePeriodSQL string
	//go:embed sql/platform/lock_snapshot.sql
	lockSnapshotSQL string
	//go:embed sql/platform/find_reservation.sql
	findReservationSQL string
	//go:embed sql/platform/lock_reservation.sql
	lockReservationSQL string
	//go:embed sql/platform/insert_decision.sql
	insertDecisionSQL string
	//go:embed sql/platform/insert_reservation.sql
	insertReservationSQL string
	//go:embed sql/platform/find_gauge_entry.sql
	findGaugeEntrySQL string
	//go:embed sql/platform/insert_gauge_entry.sql
	insertGaugeEntrySQL string
	//go:embed sql/platform/start_effect.sql
	startEffectSQL string
	//go:embed sql/platform/transition_reservation.sql
	transitionReservationSQL string
)
