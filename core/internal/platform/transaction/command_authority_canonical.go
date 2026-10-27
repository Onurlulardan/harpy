package transaction

import "slices"

func canonicalCommandGroups(
	input []CommandAuthorityGroup,
) ([]groupExpectation, map[string]string, error) {
	if len(input) < 1 || len(input) > maximumCommandGroups {
		return nil, nil, invalidCommandAuthority()
	}
	groups := append([]CommandAuthorityGroup(nil), input...)
	slices.SortFunc(groups, func(left, right CommandAuthorityGroup) int {
		return compareNames(left.ID, right.ID)
	})
	result := make([]groupExpectation, 0, len(groups))
	owners := make(map[string]string)
	previous := ""
	for _, group := range groups {
		members, exists := tenantAuthorityGroupMembers(group.ID)
		if !exists || !authorityNamePattern.MatchString(group.ID) ||
			group.Generation <= 0 || group.ID == previous {
			return nil, nil, invalidCommandAuthority()
		}
		previous = group.ID
		result = append(result, groupExpectation{
			id: group.ID, generation: group.Generation,
		})
		for _, owner := range members {
			if _, duplicate := owners[owner]; duplicate {
				return nil, nil, invalidCommandAuthority()
			}
			owners[owner] = group.ID
		}
	}
	if len(owners) < 1 || len(owners) > maximumCommandDomains {
		return nil, nil, invalidCommandAuthority()
	}
	return result, owners, nil
}

func canonicalCommandDomains(
	input []CommandAuthorityDomain,
	owners map[string]string,
) ([]domainExpectation, error) {
	if len(input) != len(owners) {
		return nil, invalidCommandAuthority()
	}
	domains := append([]CommandAuthorityDomain(nil), input...)
	slices.SortFunc(domains, func(left, right CommandAuthorityDomain) int {
		return compareNames(left.Name, right.Name)
	})
	result := make([]domainExpectation, 0, len(domains))
	previous := ""
	for _, domain := range domains {
		groupID, exists := owners[domain.Name]
		if !exists || !authorityNamePattern.MatchString(domain.Name) ||
			domain.Generation <= 0 || domain.Name == previous {
			return nil, invalidCommandAuthority()
		}
		previous = domain.Name
		result = append(result, domainExpectation{
			name: domain.Name, generation: domain.Generation, groupID: groupID,
		})
	}
	return result, nil
}

func canonicalParticipants(input []string, owners map[string]string) ([]string, error) {
	if len(input) < 1 || len(input) > len(owners) {
		return nil, invalidCommandAuthority()
	}
	participants := append([]string(nil), input...)
	slices.Sort(participants)
	previous := ""
	for _, participant := range participants {
		if _, exists := owners[participant]; !exists || participant == previous {
			return nil, invalidCommandAuthority()
		}
		previous = participant
	}
	return participants, nil
}

func compareNames(left string, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
