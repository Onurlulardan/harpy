package v1query

import "fmt"

type Dialect uint8

const (
	DialectDynamicList Dialect = iota + 1
	DialectOpenAPIGet
)

func (dialect Dialect) Validate() error {
	if dialect != DialectDynamicList && dialect != DialectOpenAPIGet {
		return fmt.Errorf("V1 REST Query dialect is invalid")
	}
	return nil
}

func (dialect Dialect) String() string {
	switch dialect {
	case DialectDynamicList:
		return "dynamic_list_v1"
	case DialectOpenAPIGet:
		return "openapi_get_v1"
	default:
		return ""
	}
}

func (Dialect) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
