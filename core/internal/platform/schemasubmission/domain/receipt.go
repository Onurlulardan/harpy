package domain

type Receipt struct {
	JobID      JobID
	Generation int64
	PlanSHA256 [32]byte
	Replay     bool
}

func NewReceipt(
	jobID JobID,
	generation int64,
	planSHA256 [32]byte,
	replay bool,
) (Receipt, bool) {
	if !jobID.Valid() || generation < 1 || planSHA256 == [32]byte{} {
		return Receipt{}, false
	}
	return Receipt{
		JobID: jobID, Generation: generation,
		PlanSHA256: planSHA256, Replay: replay,
	}, true
}
