package restqueryv1

type IssueCode string

const (
	IssueArityMismatch              IssueCode = "ARITY_MISMATCH"
	IssueASTNodeCountExceeded       IssueCode = "AST_NODE_COUNT_EXCEEDED"
	IssueBatchCountExceeded         IssueCode = "BATCH_COUNT_EXCEEDED"
	IssueBodySizeExceeded           IssueCode = "BODY_SIZE_EXCEEDED"
	IssueBooleanDepthExceeded       IssueCode = "BOOLEAN_DEPTH_EXCEEDED"
	IssueCapabilityRequired         IssueCode = "CAPABILITY_REQUIRED"
	IssueCursorSizeExceeded         IssueCode = "CURSOR_SIZE_EXCEEDED"
	IssueFieldDuplicate             IssueCode = "FIELD_DUPLICATE"
	IssueFieldForbidden             IssueCode = "FIELD_FORBIDDEN"
	IssueFieldUnknown               IssueCode = "FIELD_UNKNOWN"
	IssueFunctionNotAllowed         IssueCode = "FUNCTION_NOT_ALLOWED"
	IssueFunctionUnknown            IssueCode = "FUNCTION_UNKNOWN"
	IssueGroupCountExceeded         IssueCode = "GROUP_COUNT_EXCEEDED"
	IssueHavingCountExceeded        IssueCode = "HAVING_COUNT_EXCEEDED"
	IssueIncludeDepthExceeded       IssueCode = "INCLUDE_DEPTH_EXCEEDED"
	IssueIncludeEdgeCountExceeded   IssueCode = "INCLUDE_EDGE_COUNT_EXCEEDED"
	IssueIncludeResultCountExceeded IssueCode = "INCLUDE_RESULT_COUNT_EXCEEDED"
	IssueIndexInvalid               IssueCode = "INDEX_INVALID"
	IssueIndexNonContiguous         IssueCode = "INDEX_NON_CONTIGUOUS"
	IssueItemCountExceeded          IssueCode = "ITEM_COUNT_EXCEEDED"
	IssueLiteralCountExceeded       IssueCode = "LITERAL_COUNT_EXCEEDED"
	IssueMeasureCountExceeded       IssueCode = "MEASURE_COUNT_EXCEEDED"
	IssueMemberDuplicate            IssueCode = "MEMBER_DUPLICATE"
	IssueMemberForbidden            IssueCode = "MEMBER_FORBIDDEN"
	IssueMemberRequired             IssueCode = "MEMBER_REQUIRED"
	IssueMemberUnknown              IssueCode = "MEMBER_UNKNOWN"
	IssueNameDuplicate              IssueCode = "NAME_DUPLICATE"
	IssueNameInvalid                IssueCode = "NAME_INVALID"
	IssueNameReserved               IssueCode = "NAME_RESERVED"
	IssueNFCRequired                IssueCode = "NFC_REQUIRED"
	IssueNULForbidden               IssueCode = "NUL_FORBIDDEN"
	IssueOperandCountExceeded       IssueCode = "OPERAND_COUNT_EXCEEDED"
	IssueOperatorNotAllowed         IssueCode = "OPERATOR_NOT_ALLOWED"
	IssueOperatorUnknown            IssueCode = "OPERATOR_UNKNOWN"
	IssuePageOrdinalOverflow        IssueCode = "PAGE_ORDINAL_OVERFLOW"
	IssuePageSizeExceeded           IssueCode = "PAGE_SIZE_EXCEEDED"
	IssuePageSizeMismatch           IssueCode = "PAGE_SIZE_MISMATCH"
	IssueParameterArityMismatch     IssueCode = "PARAMETER_ARITY_MISMATCH"
	IssueParameterRequired          IssueCode = "PARAMETER_REQUIRED"
	IssueParameterUnknown           IssueCode = "PARAMETER_UNKNOWN"
	IssuePercentEncodingInvalid     IssueCode = "PERCENT_ENCODING_INVALID"
	IssuePlannerCostExceeded        IssueCode = "PLANNER_COST_EXCEEDED"
	IssuePredicateCountExceeded     IssueCode = "PREDICATE_COUNT_EXCEEDED"
	IssueQuerySizeExceeded          IssueCode = "QUERY_SIZE_EXCEEDED"
	IssueRawPlusForbidden           IssueCode = "RAW_PLUS_FORBIDDEN"
	IssueReferenceInvalid           IssueCode = "REFERENCE_INVALID"
	IssueRelationCycle              IssueCode = "RELATION_CYCLE"
	IssueRelationUnknown            IssueCode = "RELATION_UNKNOWN"
	IssueResultCountExceeded        IssueCode = "RESULT_COUNT_EXCEEDED"
	IssueSearchSizeExceeded         IssueCode = "SEARCH_SIZE_EXCEEDED"
	IssueSelectCountExceeded        IssueCode = "SELECT_COUNT_EXCEEDED"
	IssueSortCountExceeded          IssueCode = "SORT_COUNT_EXCEEDED"
	IssueTypeMismatch               IssueCode = "TYPE_MISMATCH"
	IssueUTF8Invalid                IssueCode = "UTF8_INVALID"
	IssueValueNotAllowed            IssueCode = "VALUE_NOT_ALLOWED"
	IssueValueNotCanonical          IssueCode = "VALUE_NOT_CANONICAL"
	IssueValueOutOfRange            IssueCode = "VALUE_OUT_OF_RANGE"
)

var issueCatalog = [...]IssueCode{
	IssueArityMismatch, IssueASTNodeCountExceeded, IssueBatchCountExceeded,
	IssueBodySizeExceeded, IssueBooleanDepthExceeded, IssueCapabilityRequired,
	IssueCursorSizeExceeded, IssueFieldDuplicate, IssueFieldForbidden,
	IssueFieldUnknown, IssueFunctionNotAllowed, IssueFunctionUnknown,
	IssueGroupCountExceeded, IssueHavingCountExceeded, IssueIncludeDepthExceeded,
	IssueIncludeEdgeCountExceeded, IssueIncludeResultCountExceeded, IssueIndexInvalid,
	IssueIndexNonContiguous, IssueItemCountExceeded, IssueLiteralCountExceeded,
	IssueMeasureCountExceeded, IssueMemberDuplicate, IssueMemberForbidden,
	IssueMemberRequired, IssueMemberUnknown, IssueNameDuplicate, IssueNameInvalid,
	IssueNameReserved, IssueNFCRequired, IssueNULForbidden, IssueOperandCountExceeded,
	IssueOperatorNotAllowed, IssueOperatorUnknown, IssuePageOrdinalOverflow,
	IssuePageSizeExceeded, IssuePageSizeMismatch, IssueParameterArityMismatch,
	IssueParameterRequired, IssueParameterUnknown, IssuePercentEncodingInvalid,
	IssuePlannerCostExceeded, IssuePredicateCountExceeded, IssueQuerySizeExceeded,
	IssueRawPlusForbidden, IssueReferenceInvalid, IssueRelationCycle,
	IssueRelationUnknown, IssueResultCountExceeded, IssueSearchSizeExceeded,
	IssueSelectCountExceeded, IssueSortCountExceeded, IssueTypeMismatch,
	IssueUTF8Invalid, IssueValueNotAllowed, IssueValueNotCanonical, IssueValueOutOfRange,
}

func IssueCodes() []IssueCode {
	result := make([]IssueCode, len(issueCatalog))
	copy(result, issueCatalog[:])
	return result
}

func (code IssueCode) Valid() bool {
	if !validMachineCode(string(code)) {
		return false
	}
	for _, candidate := range issueCatalog {
		if candidate == code {
			return true
		}
	}
	return false
}
