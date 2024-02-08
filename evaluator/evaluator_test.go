package evaluator

import (
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/object"
	"BigTalk_Interpreter/parser"
	"testing"
)

func TestClosure(t *testing.T) {
	input := `
let addFirst = fn(x) {
fn(y) { x + y };
};
let addSecond = addFirst(1);
addSecond(2);`
	evaluated := setupEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestEvalFunctionApplication(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}
	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func TestEvalFunctionExpression(t *testing.T) {
	input := "fn(x) { x + 1; };"

	evaluated := setupEval(input)
	fnObj, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("evaluated = %T (%+v), want object.Function", evaluated, evaluated)
	}

	if len(fnObj.Parameters) != 1 {
		t.Fatalf("len(fnObj.Parameters) = %d, want %d", len(fnObj.Parameters), 1)
	}

	if fnObj.Parameters[0].String() != "x" {
		t.Fatalf("function parameter is not 'x', got %q", fnObj.Parameters[0].String())
	}

	expectedBody := "(x + 1)"
	if fnObj.Body.String() != expectedBody {
		t.Fatalf("fnObj.Body.String() = %q, want %q", fnObj.Body.String(), expectedBody)
	}
}

func TestEvalLetStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"let x = 1; x;", 1},
		{"let x = 2 * 3; x;", 6},
		{"let x = 5; let y = x; y;", 5},
		{"let x = 5; let y = x; let z = x + y + 5; z;", 15},
	}
	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1) {
if (10 > 1) {
return true + false;
}
return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"x",
			"identifier not found: x",
		},
	}

	for _, tc := range testCases {
		evaluated := setupEval(tc.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got = %T (%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Message != tc.expected {
			t.Errorf("errObj.Message = %q, want = %q", errObj.Message, tc.expected)
		}
	}
}

func TestEvalReturnStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"return 1;", 1},
		{"return 2; 3;", 2},
		{"return 1 * 2; 3;", 2},
		{"1; return 2 * 3; 4;", 6},
		{
			`
if (3 > 1) {
	if (3 > 1) {
		return 5;
	}
	return 1;
}
`,
			5,
		},
	}

	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func TestEvalIfElseExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		expected any
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		integer, ok := tc.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestBangOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 == 1", true},
		{"1 != 1", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testBooleanObject(t, evaluated, tc.expected)
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"3", 3},
		{"12", 12},
		{"-3", -3},
		{"-12", -12},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
	}

	for _, tc := range testCases {
		evaluated := setupEval(tc.input)
		testIntegerObject(t, evaluated, tc.expected)
	}
}

func setupEval(input string) object.IObject {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.IObject, expected int64) bool {
	intObj, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("obj is not Integer, got = %T (%+v)", obj, obj)
		return false
	}

	if intObj.Value != expected {
		t.Errorf("intObj.Value = %d, want %d", intObj.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.IObject, expected bool) bool {
	boolObj, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("obj is not Boolean, got = %T (%+v)", obj, obj)
		return false
	}

	if boolObj.Value != expected {
		t.Errorf("boolObj.Value = %t, want %t", boolObj.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.IObject) bool {
	if obj != NULL {
		t.Errorf("obj = %T (%+v), want NULL.", obj, obj)
		return false
	}
	return true
}
