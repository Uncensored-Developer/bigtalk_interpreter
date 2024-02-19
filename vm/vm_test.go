package vm

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/object"
	"BigTalk_Interpreter/parser"
	"fmt"
	"testing"
)

type vmTestCase struct {
	input    string
	expected any
}

func TestVirtualMachineBooleanExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"true", true},
		{"false", false},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineIntegerArithmetic(t *testing.T) {
	testCases := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"3 * 2", 6},
		{"6 / 2", 3},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
	}
	runVirtualMachineTests(t, testCases)
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.IObject) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("actual is not *objecr.Integer. got = %T (%v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object.Value = %d, want = %d", result.Value, expected)
	}
	return nil
}

func testBooleanObject(expected bool, actual object.IObject) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("actual is not *objecr.Boolean. got = %T (%v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object.Value = %t, want = %t", result.Value, expected)
	}
	return nil
}

func runVirtualMachineTests(t *testing.T, testCases []vmTestCase) {
	t.Helper()

	for _, tc := range testCases {
		program := parse(tc.input)

		comp := compiler.NewCompiler()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		vm := NewVirtualMachine(comp.ByteCode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElement := vm.LastPoppedStackElement()

		testExpectedObject(t, tc.expected, stackElement)
	}
}

func testExpectedObject(t *testing.T, expected any, actual object.IObject) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject() failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject() failed: %s", err)
		}
	}
}
