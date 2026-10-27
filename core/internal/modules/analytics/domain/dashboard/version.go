package dashboard

type Version struct {
	id         VersionID
	dashboard  ID
	number     uint64
	revision   uint64
	published  bool
	visibility Visibility
	widgets    []Widget
}

func NewDraft(
	id VersionID, dashboard ID, number, revision uint64,
	visibility Visibility, widgets []Widget,
) (Version, error) {
	if !id.Valid() || !dashboard.Valid() || number == 0 || revision == 0 ||
		!visibility.Valid() || len(widgets) > 30 || !uniqueWidgets(widgets) {
		return Version{}, ErrInvalid
	}
	return Version{id: id, dashboard: dashboard, number: number, revision: revision,
		visibility: visibility, widgets: append([]Widget(nil), widgets...)}, nil
}

func (value Version) ID() VersionID          { return value.id }
func (value Version) DashboardID() ID        { return value.dashboard }
func (value Version) Number() uint64         { return value.number }
func (value Version) Revision() uint64       { return value.revision }
func (value Version) Visibility() Visibility { return value.visibility }
func (value Version) Published() bool        { return value.published }
func (value Version) Widgets() []Widget      { return append([]Widget(nil), value.widgets...) }

func (value Version) Publish(expected uint64) (Version, error) {
	if value.published || value.revision != expected || expected == 0 || len(value.widgets) == 0 {
		return Version{}, ErrConflict
	}
	value.published = true
	value.revision++
	return value, nil
}

func uniqueWidgets(widgets []Widget) bool {
	seenID := map[[16]byte]struct{}{}
	seenPosition := map[uint16]struct{}{}
	for _, widget := range widgets {
		if !widget.id.Valid() {
			return false
		}
		id := widget.id.Bytes()
		if _, ok := seenID[id]; ok {
			return false
		}
		if _, ok := seenPosition[widget.position]; ok {
			return false
		}
		seenID[id] = struct{}{}
		seenPosition[widget.position] = struct{}{}
	}
	return true
}
