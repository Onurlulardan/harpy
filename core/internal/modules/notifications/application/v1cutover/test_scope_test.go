package v1cutover

import "github.com/google/uuid"

func cutoverTestScope() cutoverScope {
	return cutoverScope{
		tenant: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		actor:  uuid.MustParse("10000000-0000-4000-8000-000000000002"),
	}
}
