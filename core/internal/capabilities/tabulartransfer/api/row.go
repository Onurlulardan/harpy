package api

type Row struct {
	Index  int64
	Values map[string]string
}

func (row Row) Clone() Row {
	copyValues := make(map[string]string, len(row.Values))
	for key, value := range row.Values {
		copyValues[key] = value
	}
	return Row{Index: row.Index, Values: copyValues}
}

type Reader interface {
	Columns() []string
	Next() (Row, error)
}

type Analysis struct {
	Columns []string
	Samples []Row
}
