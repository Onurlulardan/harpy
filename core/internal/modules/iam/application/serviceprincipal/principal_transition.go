package serviceprincipal

func CanTransitionPrincipal(from, to PrincipalState) bool {
	switch from {
	case PrincipalActive:
		return to == PrincipalDisabled || to == PrincipalRetired
	case PrincipalDisabled:
		return to == PrincipalActive || to == PrincipalRetired
	default:
		return false
	}
}
