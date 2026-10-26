package csrf

type Credential struct {
	wire   []byte
	active bool
}

func (credential *Credential) RevealOnce(consume func([]byte) error) error {
	if credential == nil || !credential.active || consume == nil {
		return ErrInvalidCredential
	}
	err := consume(credential.wire)
	credential.Destroy()
	return err
}

func (credential *Credential) Destroy() {
	if credential == nil {
		return
	}
	wipe(credential.wire)
	credential.wire = nil
	credential.active = false
}

func wipe(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
