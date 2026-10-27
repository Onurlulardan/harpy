package audithttp

import "encoding/json"

func marshal(value any) ([]byte, error) { return json.Marshal(value) }
