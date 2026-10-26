package api

type SurfaceKind string

const (
	SurfaceRecordCreate SurfaceKind = "record_create"
	SurfaceRecordUpdate SurfaceKind = "record_update"
	SurfaceRecordDetail SurfaceKind = "record_detail"
	SurfaceRecordList   SurfaceKind = "record_list"
	SurfaceRecordFilter SurfaceKind = "record_filter"
	SurfacePrint        SurfaceKind = "print"
)

func (kind SurfaceKind) Valid() bool {
	switch kind {
	case SurfaceRecordCreate, SurfaceRecordUpdate, SurfaceRecordDetail,
		SurfaceRecordList, SurfaceRecordFilter, SurfacePrint:
		return true
	default:
		return false
	}
}

func (kind SurfaceKind) Required() bool {
	return kind.Valid() && kind != SurfacePrint
}

type SurfaceDefinition struct {
	Kind         SurfaceKind   `json:"kind"`
	Sections     []Section     `json:"sections"`
	Translations []Translation `json:"translations"`
}
