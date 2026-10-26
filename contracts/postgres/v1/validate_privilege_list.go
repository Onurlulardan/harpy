package postgresv1

import "strings"

var privilegeOrder = []string{
	"CONNECT", "CREATE", "TEMP", "USAGE", "SELECT", "INSERT", "UPDATE",
	"DELETE", "TRUNCATE", "REFERENCES", "TRIGGER", "EXECUTE",
}

func validatePrivilegeList(kind string, privileges []string) bool {
	if privileges == nil || len(privileges) == 0 {
		return false
	}
	allowed := allowedPrivileges(kind)
	previous := -1
	for _, privilege := range privileges {
		rank := privilegeRank(privilege)
		if rank <= previous || !strings.Contains("|"+allowed+"|", "|"+privilege+"|") {
			return false
		}
		previous = rank
	}
	return true
}

func allowedPrivileges(kind string) string {
	switch kind {
	case "database":
		return "CONNECT|CREATE|TEMP"
	case "schema":
		return "CREATE|USAGE"
	case "table", "partitioned_table", "view", "materialized_view":
		return "SELECT|INSERT|UPDATE|DELETE|TRUNCATE|REFERENCES|TRIGGER"
	case "sequence":
		return "USAGE|SELECT|UPDATE"
	case "function", "procedure":
		return "EXECUTE"
	case "type":
		return "USAGE"
	default:
		return ""
	}
}

func privilegeRank(value string) int {
	for index, privilege := range privilegeOrder {
		if value == privilege {
			return index
		}
	}
	return -1
}

func aclGrantKey(grant ACLGrant) string {
	return strings.Join([]string{
		CatalogObjectKey(grant.Object), grant.Grantor, grant.Grantee,
		strings.Join(grant.Privileges, ","),
	}, "|")
}
