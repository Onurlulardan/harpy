package keywrap

import (
	"errors"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const formatID = "FCKEY1"

var (
	ErrInvalidWrap = errors.New("documents wrapped key format is invalid")
	ErrKeyTampered = errors.New("documents wrapped key authentication failed")
)

type Context struct {
	TenantID, DocumentID, VersionID domain.ID
	KeyID                           string
}

func (context Context) valid() bool {
	return context.TenantID.Valid() && context.DocumentID.UUIDv7() &&
		context.VersionID.UUIDv7() && validKeyID(context.KeyID)
}

func validKeyID(value string) bool {
	if len(value) < 1 || len(value) > 64 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if character < 'a' || character > 'z' {
			if character < '0' || character > '9' {
				if character != '-' && character != '_' && character != '.' {
					return false
				}
			}
		}
	}
	return true
}

func header(context Context) ([]byte, error) {
	if !context.valid() {
		return nil, ErrInvalidWrap
	}
	result := make([]byte, 6+1+len(context.KeyID)+48)
	copy(result, formatID)
	result[6] = byte(len(context.KeyID))
	copy(result[7:], context.KeyID)
	offset := 7 + len(context.KeyID)
	copy(result[offset:offset+16], context.TenantID[:])
	copy(result[offset+16:offset+32], context.DocumentID[:])
	copy(result[offset+32:offset+48], context.VersionID[:])
	return result, nil
}

func decodeHeader(value []byte) (Context, []byte, int, error) {
	if len(value) < 6+1+1+48+12+48 || string(value[:6]) != formatID {
		return Context{}, nil, 0, ErrInvalidWrap
	}
	keyLength := int(value[6])
	headerLength := 7 + keyLength + 48
	if keyLength == 0 || len(value) != headerLength+12+48 {
		return Context{}, nil, 0, ErrInvalidWrap
	}
	context := Context{KeyID: string(value[7 : 7+keyLength])}
	offset := 7 + keyLength
	copy(context.TenantID[:], value[offset:offset+16])
	copy(context.DocumentID[:], value[offset+16:offset+32])
	copy(context.VersionID[:], value[offset+32:offset+48])
	if !context.valid() {
		return Context{}, nil, 0, ErrInvalidWrap
	}
	return context, value[:headerLength], headerLength, nil
}
