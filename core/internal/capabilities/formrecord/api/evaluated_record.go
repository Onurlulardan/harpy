package recordkernel

type EvaluatedRecord struct {
	record       Record
	canonical    []byte
	instructions uint64
}

func (result EvaluatedRecord) Record() Record { return result.record.clone() }

func (result EvaluatedRecord) CanonicalJSON() []byte {
	return append([]byte(nil), result.canonical...)
}

func (result EvaluatedRecord) Instructions() uint64 { return result.instructions }
