package fcartv1

import "fmt"

type KEK struct {
	id       string
	material [DEKBytes]byte
}

func NewKEK(id string, material []byte) (KEK, error) {
	if !validKeyID(id) || len(material) != DEKBytes {
		return KEK{}, ErrInvalidKEK
	}
	key := KEK{id: id}
	copy(key.material[:], material)
	var nonzero byte
	for _, value := range key.material {
		nonzero |= value
	}
	if nonzero == 0 {
		return KEK{}, ErrInvalidKEK
	}
	return key, nil
}

func (key KEK) ID() string   { return key.id }
func (KEK) String() string   { return "[REDACTED FCART1 KEK]" }
func (KEK) GoString() string { return "[REDACTED FCART1 KEK]" }
func (KEK) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("[REDACTED FCART1 KEK]"))
}
