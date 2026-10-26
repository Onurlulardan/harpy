package query

type Aggregate struct {
	id             ID
	lifecycle      Lifecycle
	rowRevision    uint64
	referenceEpoch uint64
	currentDraft   VersionID
}

func Rehydrate(
	id ID, lifecycle Lifecycle, rowRevision, referenceEpoch uint64,
	currentDraft VersionID,
) (Aggregate, error) {
	if !id.Valid() || !lifecycle.Valid() || rowRevision == 0 || referenceEpoch == 0 ||
		(lifecycle == LifecycleActive && !currentDraft.Valid()) {
		return Aggregate{}, ErrInvalid
	}
	return Aggregate{id: id, lifecycle: lifecycle, rowRevision: rowRevision,
		referenceEpoch: referenceEpoch, currentDraft: currentDraft}, nil
}

func (value Aggregate) ID() ID                  { return value.id }
func (value Aggregate) Lifecycle() Lifecycle    { return value.lifecycle }
func (value Aggregate) RowRevision() uint64     { return value.rowRevision }
func (value Aggregate) ReferenceEpoch() uint64  { return value.referenceEpoch }
func (value Aggregate) CurrentDraft() VersionID { return value.currentDraft }

func (value Aggregate) Retire(expected uint64) (Aggregate, error) {
	if value.lifecycle != LifecycleActive || expected == 0 || expected != value.rowRevision {
		return Aggregate{}, ErrConflict
	}
	value.lifecycle = LifecycleRetired
	value.rowRevision++
	value.referenceEpoch++
	value.currentDraft = VersionID{}
	return value, nil
}
