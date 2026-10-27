package transaction

func validCommandOperation(operation VerifiedOperation) bool {
	if len(operation.groups) < 1 || len(operation.groups) > maximumCommandGroups ||
		len(operation.domains) < 1 || len(operation.domains) > maximumCommandDomains ||
		len(operation.participants) < 1 {
		return false
	}
	owners, valid := validOperationGroups(operation.groups)
	if !valid || len(owners) != len(operation.domains) {
		return false
	}
	if !validOperationDomains(operation.domains, owners) {
		return false
	}
	return validOperationParticipants(operation.participants, owners)
}

func validOperationGroups(groups []groupExpectation) (map[string]string, bool) {
	owners := make(map[string]string)
	previous := ""
	for _, group := range groups {
		members, exists := tenantAuthorityGroupMembers(group.id)
		if !exists || group.generation <= 0 || group.id <= previous {
			return nil, false
		}
		previous = group.id
		for _, owner := range members {
			if _, duplicate := owners[owner]; duplicate {
				return nil, false
			}
			owners[owner] = group.id
		}
	}
	return owners, len(owners) <= maximumCommandDomains
}

func validOperationDomains(domains []domainExpectation, owners map[string]string) bool {
	previous := ""
	for _, domain := range domains {
		groupID, exists := owners[domain.name]
		if !exists || domain.generation <= 0 || domain.name <= previous ||
			domain.groupID != groupID {
			return false
		}
		previous = domain.name
	}
	return true
}

func validOperationParticipants(participants []string, owners map[string]string) bool {
	if len(participants) > len(owners) {
		return false
	}
	previous := ""
	for _, participant := range participants {
		if _, exists := owners[participant]; !exists || participant <= previous {
			return false
		}
		previous = participant
	}
	return true
}
