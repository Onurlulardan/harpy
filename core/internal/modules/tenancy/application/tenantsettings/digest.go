package tenantsettings

import (
	"crypto/sha256"
	"encoding/binary"
)

func updateDigest(command UpdateCommand) [32]byte {
	document := []byte("harpy/tenancy/settings-update/v1")
	document = binary.BigEndian.AppendUint64(document, uint64(command.ExpectedRevision))
	if command.Name != nil {
		document = appendField(document, []byte(*command.Name))
	} else {
		document = appendField(document, nil)
	}
	if command.SMTP != nil {
		document = appendField(document, []byte(command.SMTP.Host))
		document = binary.BigEndian.AppendUint32(document, uint32(command.SMTP.Port))
		document = appendField(document, []byte(command.SMTP.Username))
		document = appendField(document, command.SMTP.Password)
		document = appendField(document, []byte(command.SMTP.From))
	} else {
		document = appendField(document, nil)
	}
	result := sha256.Sum256(document)
	clear(document)
	return result
}

func testEmailDigest(command TestEmailCommand) [32]byte {
	document := []byte("harpy/tenancy/settings-test-email/v1")
	document = binary.BigEndian.AppendUint64(document, uint64(command.ExpectedRevision))
	document = appendField(document, []byte(command.To))
	return sha256.Sum256(document)
}

func appendField(document, value []byte) []byte {
	document = binary.BigEndian.AppendUint32(document, uint32(len(value)))
	return append(document, value...)
}
