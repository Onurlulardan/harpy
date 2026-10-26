package restqueryv1

type ErrorCode string

type ErrorDefinition struct {
	Code       ErrorCode
	HTTPStatus int
}

const (
	ErrorQueryAggregateOverflow       ErrorCode = "QUERY_AGGREGATE_OVERFLOW"
	ErrorQueryBudgetExceeded          ErrorCode = "QUERY_BUDGET_EXCEEDED"
	ErrorQueryCountBudgetExceeded     ErrorCode = "QUERY_COUNT_BUDGET_EXCEEDED"
	ErrorQueryCountTimeout            ErrorCode = "QUERY_COUNT_TIMEOUT"
	ErrorQueryCursorExpired           ErrorCode = "QUERY_CURSOR_EXPIRED"
	ErrorQueryCursorInvalid           ErrorCode = "QUERY_CURSOR_INVALID"
	ErrorQueryCursorStale             ErrorCode = "QUERY_CURSOR_STALE"
	ErrorQueryDuplicateMember         ErrorCode = "QUERY_DUPLICATE_MEMBER"
	ErrorQueryFieldForbidden          ErrorCode = "QUERY_FIELD_FORBIDDEN"
	ErrorQueryForbidden               ErrorCode = "QUERY_FORBIDDEN"
	ErrorQueryInvalid                 ErrorCode = "QUERY_INVALID"
	ErrorQueryParameterInvalid        ErrorCode = "QUERY_PARAMETER_INVALID"
	ErrorQueryTimeout                 ErrorCode = "QUERY_TIMEOUT"
	ErrorQueryTransportInvalid        ErrorCode = "QUERY_TRANSPORT_INVALID"
	ErrorRelationCardinalityViolation ErrorCode = "RELATION_CARDINALITY_VIOLATION"
)

var errorCatalog = [...]ErrorDefinition{
	{Code: ErrorQueryAggregateOverflow, HTTPStatus: 422},
	{Code: ErrorQueryBudgetExceeded, HTTPStatus: 422},
	{Code: ErrorQueryCountBudgetExceeded, HTTPStatus: 422},
	{Code: ErrorQueryCountTimeout, HTTPStatus: 504},
	{Code: ErrorQueryCursorExpired, HTTPStatus: 400},
	{Code: ErrorQueryCursorInvalid, HTTPStatus: 400},
	{Code: ErrorQueryCursorStale, HTTPStatus: 409},
	{Code: ErrorQueryDuplicateMember, HTTPStatus: 422},
	{Code: ErrorQueryFieldForbidden, HTTPStatus: 403},
	{Code: ErrorQueryForbidden, HTTPStatus: 403},
	{Code: ErrorQueryInvalid, HTTPStatus: 422},
	{Code: ErrorQueryParameterInvalid, HTTPStatus: 422},
	{Code: ErrorQueryTimeout, HTTPStatus: 504},
	{Code: ErrorQueryTransportInvalid, HTTPStatus: 400},
	{Code: ErrorRelationCardinalityViolation, HTTPStatus: 409},
}

func ErrorCatalog() []ErrorDefinition {
	result := make([]ErrorDefinition, len(errorCatalog))
	copy(result, errorCatalog[:])
	return result
}

func (code ErrorCode) Valid() bool {
	if !validMachineCode(string(code)) {
		return false
	}
	_, found := ErrorStatus(code)
	return found
}

func ErrorStatus(code ErrorCode) (int, bool) {
	for _, definition := range errorCatalog {
		if definition.Code == code {
			return definition.HTTPStatus, true
		}
	}
	return 0, false
}
