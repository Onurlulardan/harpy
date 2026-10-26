package ddl

import (
	"encoding/json"
	"fmt"
	"strings"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func RenderTypedLiteralV2(literal schemav2.TypedLiteral) (string, error) {
	if err := schemav2.ValidateTypedLiteralV2(&literal); err != nil {
		return "", err
	}
	switch literal.StorageType {
	case schemav2.StorageText, schemav2.StorageEnum:
		return renderStringLiteral(literal.Value, "text")
	case schemav2.StorageInteger:
		return renderStringLiteral(literal.Value, "bigint")
	case schemav2.StorageDecimal:
		return renderStringLiteral(literal.Value, "numeric")
	case schemav2.StorageBoolean:
		if string(literal.Value) == "true" {
			return "TRUE", nil
		}
		return "FALSE", nil
	case schemav2.StorageDate:
		return renderStringLiteral(literal.Value, "date")
	case schemav2.StorageTimestamp:
		return renderStringLiteral(literal.Value, "timestamp with time zone")
	case schemav2.StorageUUID:
		return renderStringLiteral(literal.Value, "uuid")
	case schemav2.StorageTextList:
		return renderArrayLiteral(literal.Value, "text")
	case schemav2.StorageUUIDList:
		return renderArrayLiteral(literal.Value, "uuid")
	case schemav2.StorageJSON:
		return quoteLiteral(string(literal.Value)) + "::jsonb", nil
	default:
		return "", fmt.Errorf("typed literal storage is not renderable")
	}
}

func renderStringLiteral(raw []byte, cast string) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("typed string literal is invalid")
	}
	return quoteLiteral(value) + "::" + cast, nil
}

func renderArrayLiteral(raw []byte, cast string) (string, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return "", fmt.Errorf("typed array literal is invalid")
	}
	elements := make([]string, len(values))
	for index, value := range values {
		elements[index] = quoteLiteral(value)
	}
	return "ARRAY[" + strings.Join(elements, ",") + "]::" + cast + "[]", nil
}

func quoteLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
