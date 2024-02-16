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

func TestIntegerArithmetic(t *testing.T) {
	testCases := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.MakeInstruction(code.OpConstant, 0),
				code.MakeInstruction(code.OpConstant, 1),
				code.MakeInstruction(code.OppAdd),
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
