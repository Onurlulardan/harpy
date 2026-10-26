package ddl

type Statement struct {
	Text               string
	Arguments          []any
	RecordPageEvidence bool
}
