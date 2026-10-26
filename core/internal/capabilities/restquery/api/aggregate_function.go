package api

type AggregateFunction uint8

const (
	AggregateFunctionCount AggregateFunction = iota + 1
	AggregateFunctionCountDistinct
	AggregateFunctionSum
	AggregateFunctionAvg
	AggregateFunctionMin
	AggregateFunctionMax
)

func ParseAggregateFunction(value string) (AggregateFunction, bool) {
	switch value {
	case "count":
		return AggregateFunctionCount, true
	case "count_distinct":
		return AggregateFunctionCountDistinct, true
	case "sum":
		return AggregateFunctionSum, true
	case "avg":
		return AggregateFunctionAvg, true
	case "min":
		return AggregateFunctionMin, true
	case "max":
		return AggregateFunctionMax, true
	default:
		return 0, false
	}
}

func (function AggregateFunction) String() string {
	switch function {
	case AggregateFunctionCount:
		return "count"
	case AggregateFunctionCountDistinct:
		return "count_distinct"
	case AggregateFunctionSum:
		return "sum"
	case AggregateFunctionAvg:
		return "avg"
	case AggregateFunctionMin:
		return "min"
	case AggregateFunctionMax:
		return "max"
	default:
		return ""
	}
}

func (function AggregateFunction) valid() bool {
	return function >= AggregateFunctionCount && function <= AggregateFunctionMax
}

func (AggregateFunction) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
