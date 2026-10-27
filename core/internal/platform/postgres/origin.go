package postgres

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func originForVectorSchema(schema string) (runtimev1.DatabaseOrigin, error) {
	switch schema {
	case "platform_ext":
		return runtimev1.DatabaseOriginFresh, nil
	case "public":
		return runtimev1.DatabaseOriginV1Adopted, nil
	default:
		return "", fmt.Errorf("%w: vector schema", ErrStartupPreflight)
	}
}
