package resolved

import (
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func sameFieldType(left, right api.FieldType) bool {
	leftKind, leftErr := left.Kind()
	rightKind, rightErr := right.Kind()
	if leftErr != nil || rightErr != nil || leftKind != rightKind {
		return false
	}
	switch leftKind {
	case api.LogicalTypeText, api.LogicalTypeTextList:
		leftProfile, leftFound, leftProfileErr := left.Text()
		rightProfile, rightFound, rightProfileErr := right.Text()
		return leftProfileErr == nil && rightProfileErr == nil &&
			leftFound && rightFound && leftProfile == rightProfile
	case api.LogicalTypeInteger:
		leftProfile, leftFound, leftProfileErr := left.Integer()
		rightProfile, rightFound, rightProfileErr := right.Integer()
		return leftProfileErr == nil && rightProfileErr == nil &&
			leftFound && rightFound && leftProfile == rightProfile
	case api.LogicalTypeDecimal:
		leftProfile, leftFound, leftProfileErr := left.Decimal()
		rightProfile, rightFound, rightProfileErr := right.Decimal()
		return leftProfileErr == nil && rightProfileErr == nil &&
			leftFound && rightFound && leftProfile == rightProfile
	case api.LogicalTypeEnum:
		return sameEnumType(left, right)
	default:
		return true
	}
}

func sameEnumType(left, right api.FieldType) bool {
	leftProfile, leftFound, leftErr := left.Enum()
	rightProfile, rightFound, rightErr := right.Enum()
	if leftErr != nil || rightErr != nil || !leftFound || !rightFound {
		return false
	}
	leftVersion, leftVersionErr := leftProfile.Version()
	rightVersion, rightVersionErr := rightProfile.Version()
	leftOrdered, leftOrderedErr := leftProfile.Ordered()
	rightOrdered, rightOrderedErr := rightProfile.Ordered()
	leftValues, leftValuesErr := leftProfile.Values()
	rightValues, rightValuesErr := rightProfile.Values()
	return leftVersionErr == nil && rightVersionErr == nil &&
		leftOrderedErr == nil && rightOrderedErr == nil &&
		leftValuesErr == nil && rightValuesErr == nil &&
		leftVersion == rightVersion && leftOrdered == rightOrdered &&
		slices.Equal(leftValues, rightValues)
}
