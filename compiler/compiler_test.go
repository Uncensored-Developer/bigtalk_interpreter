package compiler

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/object"
	"BigTalk_Interpreter/parser"
	"fmt"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code.Instructions
}

func TestCompileStringExpressions(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input:             `"foobar"`,
			expectedConstants: []interface{}{"foobar"},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             `"foo" + "bar"`,
			expectedConstants: []interface{}{"foo", "bar"},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, testCases)
}

func TestCompileGlobalLetStatements(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input: `
let x = 1;
let y = 2;
`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpSetGlobal, 1),
			},
		},
		{
			input: `
let x = 1;
x;
`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpGetGlobal, 0),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input: `
let x = 1;
let y = x;
y;
`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpSetGlobal, 0),
				code.MakeInstruction(code.OpGetGlobal, 0),
				code.MakeInstruction(code.OpSetGlobal, 1),
				code.MakeInstruction(code.OpGetGlobal, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, testCases)
}

func TestCompileConditionals(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input:             "if (true) { 10 }; 20;",
			expectedConstants: []any{10, 20},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),              // 0000
				code.MakeInstruction(code.OpJumpNotTruthy, 10), // 0001
				code.MakeInstruction(code.OpConstant, 0),       // 0004
				code.MakeInstruction(code.OpJump, 11),          // 0007
				code.MakeInstruction(code.OpNull),              // 0010
				code.MakeInstruction(code.OpPop),               // 0011
				code.MakeInstruction(code.OpConstant, 1),       // 0012
				code.MakeInstruction(code.OpPop),               // 0015
			},
		},
		{
			input:             "if (true) { 10 } else { 20 }; 30;",
			expectedConstants: []any{10, 20, 30},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),              // 0000
				code.MakeInstruction(code.OpJumpNotTruthy, 10), // 0001
				code.MakeInstruction(code.OpConstant, 0),       // 0004
				code.MakeInstruction(code.OpJump, 13),          // 0007
				code.MakeInstruction(code.OpConstant, 1),       // 0010
				code.MakeInstruction(code.OpPop),               // 0013
				code.MakeInstruction(code.OpConstant, 2),       // 0014
				code.MakeInstruction(code.OpPop),               // 0014
			},
		},
	}
	runCompilerTests(t, testCases)
}

func TestCompileBooleanExpressions(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpNotEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpFalse),
				code.MakeInstruction(code.OpNotEqual),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpGreaterThan),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpGreaterThan),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpTrue),
				code.MakeInstruction(code.OpBang),
				code.MakeInstruction(code.OpPop),
			},
		},
	}
	runCompilerTests(t, testCases)
}

func TestCompileIntegerArithmetic(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpAdd),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpPop),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpSub),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "2 * 2",
			expectedConstants: []any{2, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpMul),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "2 / 2",
			expectedConstants: []any{2, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OpDiv),
				code.MakeInstruction(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpMinus),
				code.MakeInstruction(code.OpPop),
			},
		},
	}

	runCompilerTests(t, testCases)
}

func parse(input string) *ast.Program {
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	return p.ParseProgram()
}

func runCompilerTests(t *testing.T, testCases []compilerTestCase) {
	t.Helper()

	for _, tc := range testCases {
		program := parse(tc.input)

		compiler := NewCompiler()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.ByteCode()

		err = testInstructions(tc.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions() error: %s", err)
		}

		err = testConstants(t, tc.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testInstructions() error: %s", err)
		}
	}

}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatenated := concatInstructions(expected)

	if len(actual) != len(concatenated) {
		return fmt.Errorf("wrong instruction length. got = %q, want = %q", actual, concatenated)
	}
	for i, ins := range concatenated {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d. got = %q, want = %q", i, actual[i], ins)
		}
	}
	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testConstants(t *testing.T, expected []any, actual []object.IObject) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("wrong instruction length. got = %q, want = %q", len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("testIntegerObject for constant %d failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("testStringObject() failed for constant %d: %s", i, err)
			}
		}
	}
	return nil
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
