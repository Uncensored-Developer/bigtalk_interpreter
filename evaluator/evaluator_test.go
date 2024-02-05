package evaluator

import (
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/object"
	"BigTalk_Interpreter/parser"
	"testing"
)

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
	return Eval(program)
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
