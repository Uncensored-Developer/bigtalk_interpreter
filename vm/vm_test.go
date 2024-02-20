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

func TestVirtualMachineArrayLiteral(t *testing.T) {
	testCases := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"foobar"`, "foobar"},
		{`"foo" + "bar"`, "foobar"},
		{`"foo" + "bar" + "baz"`, "foobarbaz"},
	}
	runVirtualMachineTests(t, tests)
}

func TestVirtualMachineGlobalLetStatements(t *testing.T) {
	testCases := []vmTestCase{
		{"let x = 1; x", 1},
		{"let x = 1; let y = 2; x + y", 3},
		{"let x = 1; let y = x + x; x + y", 3},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineConditionals(t *testing.T) {
	testCases := []vmTestCase{
		{"if (true) { 1 }", 1},
		{"if (true) { 1 } else { 2 }", 1},
		{"if (false) { 1 } else { 2 } ", 2},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineBooleanExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"true", true},
		{"false", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!3", false},
		{"!!true", true},
		{"!!false", false},
		{"!!3", true},
		{"!(if (false) { 5; })", true},
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
		{"-1", -1},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject() failed: %s", err)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("actual not *object.Array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Items) != len(expected) {
			t.Errorf("len(array.Items) = %d, want = %d", len(array.Items), len(expected))
		}

		for i, expectedItem := range expected {
			err := testIntegerObject(int64(expectedItem), array.Items[i])
			if err != nil {
				t.Errorf("testIntegerObject() failed: %s", err)
			}
		}
	}
}

func testStringObject(expected string, actual object.IObject) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("actual is not *objecr.String. got = %T (%v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object.Value = %s, want = %s", result.Value, expected)
	}
	return nil
}
