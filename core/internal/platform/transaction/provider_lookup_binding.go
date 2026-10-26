package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

func ProviderLookupBinding(lookup ProviderLookup) [32]byte {
	if !lookup.valid() {
		return [32]byte{}
	}
	return providerLookupBinding(lookup)
}

func providerLookupBinding(lookup ProviderLookup) [32]byte {
	var buffer bytes.Buffer
	values := []string{
		"billing-provider-target.v1", lookup.Provider(), lookup.AccountID(),
		lookup.ObjectKind(), lookup.ObjectID(), lookup.EventID(),
	}
	for _, value := range values {
		_ = binary.Write(&buffer, binary.BigEndian, uint32(len(value)))
		_, _ = buffer.WriteString(value)
	}
	authentication := lookup.AuthenticationSHA256()
	_, _ = buffer.Write(authentication[:])
	return sha256.Sum256(buffer.Bytes())
}
