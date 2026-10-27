package recordkernel

type evaluationFrame struct {
	node    expression
	entered bool
	next    int
	values  []Value
}

func evaluateExpression(
	root expression,
	record Record,
	targetDecimal *DecimalProfile,
	budget *instructionBudget,
) (Value, error) {
	stack := []evaluationFrame{{node: root}}
	for len(stack) > 0 {
		index := len(stack) - 1
		frame := &stack[index]
		if !frame.entered {
			if err := budget.consume(); err != nil {
				return Value{}, err
			}
			frame.entered = true
			if frame.node.kind == "literal" || frame.node.kind == "field" {
				value := evaluateLeaf(frame.node, record)
				stack = stack[:index]
				if len(stack) == 0 {
					return value, nil
				}
				stack[len(stack)-1].values = append(stack[len(stack)-1].values, value)
				continue
			}
		}
		if frame.node.kind == "call" && frame.node.function == "coalesce" &&
			len(frame.values) > 0 && !frame.values[len(frame.values)-1].null {
			value := frame.values[len(frame.values)-1].clone()
			stack = stack[:index]
			if len(stack) == 0 {
				return value, nil
			}
			stack[len(stack)-1].values = append(stack[len(stack)-1].values, value)
			continue
		}
		if frame.next < len(frame.node.arguments) {
			child := frame.node.arguments[frame.next]
			frame.next++
			stack = append(stack, evaluationFrame{node: child})
			continue
		}
		value, err := evaluateNode(frame.node, frame.values, targetDecimal)
		if err != nil {
			return Value{}, err
		}
		stack = stack[:index]
		if len(stack) == 0 {
			return value, nil
		}
		stack[len(stack)-1].values = append(stack[len(stack)-1].values, value)
	}
	return Value{}, ErrCalculation
}

func evaluateLeaf(node expression, record Record) Value {
	if node.kind == "literal" {
		return node.literal.clone()
	}
	value, found := record.Value(node.fieldID)
	if !found {
		return nullValue(node.typeID)
	}
	return value
}

func evaluateNode(
	node expression,
	values []Value,
	targetDecimal *DecimalProfile,
) (Value, error) {
	if node.kind == "call" && node.function == "coalesce" {
		return nullValue(node.typeID), nil
	}
	for _, value := range values {
		if value.null {
			return nullValue(node.typeID), nil
		}
	}
	switch node.kind {
	case "unary":
		return evaluateUnary(node, values[0], targetDecimal)
	case "binary":
		return evaluateBinary(node, values[0], values[1], targetDecimal)
	case "call":
		return evaluateCall(node, values, targetDecimal)
	default:
		return Value{}, ErrCalculation
	}
}
