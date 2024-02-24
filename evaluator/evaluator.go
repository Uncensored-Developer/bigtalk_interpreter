package evaluator

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/object"
	"fmt"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Eval evaluates the abstract syntax tree (AST) node and returns its computed value or an error.
func Eval(node ast.INode, env *object.Environment) object.IObject {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Value, env)
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Value, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.LeftValue, env)
		if isError(left) {
			return left
		}

		right := Eval(node.RightValue, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	case *ast.CallExpression:
		function := Eval(node.Func, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		items := evalExpressions(node.Items, env)
		if len(items) == 1 && isError(items[0]) {
			return items[0]
		}
		return &object.Array{Items: items}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *ast.MapLiteral:
		return evalMapLiteral(node, env)
	}
	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.IObject {
	var result object.IObject
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.IObject) object.IObject {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.IObject) object.IObject {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.IObject) object.IObject {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// evalInfixExpression evaluates the given infix expression node and returns the result of the computation.
// It supports various operators including arithmetic and comparison operators for integers and string concatenation for strings.
// If the operator and operands have incompatible types, it returns an error.
// Supported operator: "+", "-", "*", "/", "==", "!=", "<", ">", "+"
// Parameters: operator (string) - the operator to evaluate
//
//	left (object.IObject) - the left operand
//	right (object.IObject) - the right operand
//
// Returns: object.IObject - the result of the computation or an error
func evalInfixExpression(operator string, left, right object.IObject) object.IObject {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.IObject) object.IObject {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.IObject) object.IObject {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.IObject {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(obj object.IObject) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.IObject {
	var result object.IObject
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			resultType := result.Type()
			if resultType == object.RETURN_VALUE_OBJ || resultType == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.IObject {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalExpressions(exps []ast.IExpression, env *object.Environment) []object.IObject {
	var results []object.IObject

	for _, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []object.IObject{evaluated}
		}
		results = append(results, evaluated)
	}

	return results
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.IObject) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// applyFunction applies a function object to a list of arguments and returns the result.
// If the function object is of type *object.Function, it creates an extended environment by
// setting the function's parameters as keys and the arguments as values, and then evaluates
// the function's body in this extended environment. The result is then unwrapped, removing
// any return value wrapper.
// If the function object is of type *object.Builtin, it simply calls the builtin function
// passing the arguments as arguments.
// If the function object is of any other type, it returns an error object indicating that
// the object is not a function.
func applyFunction(fn object.IObject, args []object.IObject) object.IObject {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendedFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if result := fn.Fn(args...); result != nil {
			return result
		}
		return NULL
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendedFunctionEnv(fn *object.Function, args []object.IObject) *object.Environment {
	env := object.NewWrappedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func unwrapReturnValue(obj object.IObject) object.IObject {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index object.IObject) object.IObject {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.MAP_OBJ:
		return evalMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.IObject) object.IObject {
	arrayObj := array.(*object.Array)
	idx := index.(*object.Integer).Value
	maxIndex := int64(len(arrayObj.Items) - 1)

	if idx < 0 || idx > maxIndex {
		return NULL
	}

	return arrayObj.Items[idx]
}

// evalMapLiteral evaluates the map literal AST node and returns a Map object.
// It iterates over the key-value pairs in the node and evaluates each key and value.
// If there is an error during evaluation, it returns the error.
// If a key is not hashable, it returns an error.
// It then creates a MapPair with the evaluated key and value, and adds it to the pairs map.
// Finally, it returns a new Map object with the computed pairs.
func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.IObject {
	pairs := make(map[object.HashKey]object.MapPair)

	for k, v := range node.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.IHashable)
		if !ok {
			return newError("unusable hash key: %s", key.Type())
		}

		value := Eval(v, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.MapPair{Key: key, Value: value}
	}

	return &object.Map{Pairs: pairs}
}

func evalMapIndexExpression(hashMap, index object.IObject) object.IObject {
	mapObj := hashMap.(*object.Map)

	key, ok := index.(object.IHashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := mapObj.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}
