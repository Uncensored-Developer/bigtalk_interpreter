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

func TestVirtualMachineClosures(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let firstClosure = fn(a) {
				fn() { a; };
			};

			let closure = firstClosure(3);
			closure();
			`,
			expected: 3,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				fn(c) { a + b + c };
			};

			let adder = newAdder(1, 2);
			adder(3);
			`,
			expected: 6,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				let c = a + b;
				fn(d) { c + d };
			};

			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `
			let newAdderOuter = fn(a, b) {
				let c = a + b;
				fn(d) {
					let e = d + c;
					fn(f) { e + f; };
				};
			};

			let newAdderInner = newAdderOuter(1, 2)
			let adder = newAdderInner(3);
			adder(4);
			`,
			expected: 10,
		},
		{
			input: `
			let a = 1;
			let newAdderOuter = fn(b) {
				fn(c) {
					fn(d) { a + b + c + d };
				};
			};

			let newAdderInner = newAdderOuter(2)
			let adder = newAdderInner(3);
			adder(4);
			`,
			expected: 10,
		},
		{
			input: `
			let newClosure = fn(a, b) {
				let one = fn() { a; };
				let two = fn() { b; };
				fn() { one() + two(); };
			};

			let closure = newClosure(1, 2);
			closure();
			`,
			expected: 3,
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestNewVirtualMachineBuiltinFunctions(t *testing.T) {
	testCases := []vmTestCase{
		{`len("")`, 0},
		{`len("two")`, 3},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{`len("one", "two")`,
			&object.Error{
				Message: "wrong number of arguments. got=2, want=1",
			},
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`print("hello", "world!")`, Null},
		{`tail([1, 2, 3])`, []int{2, 3}},
		{`tail([])`, Null},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be ARRAY, got INTEGER",
			},
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineCallingFunctionsWithWrongArgs(t *testing.T) {
	testCases := []vmTestCase{
		{
			input:    `fn() { 1; }(1);`,
			expected: `wrong number of arguments: got = 0, want = 1`,
		},
		{
			input:    `fn(a) { a; }();`,
			expected: `wrong number of arguments: got = 1, want = 0`,
		},
		{
			input:    `fn(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: got = 2, want = 1`,
		},
	}

	for _, tc := range testCases {
		program := parse(tc.input)

		comp := compiler.NewCompiler()
		err := comp.Compile(program)

		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		vm := NewVirtualMachine(comp.ByteCode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("Expected error from VirtualMachine but got nil")
		}

		if err.Error() != tc.expected {
			t.Fatalf("err.Error() = %q, want = %q", err, tc.expected)
		}
	}
}

func TestVirtualMachineCallingFunctionsWithArgsAndBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
let x = fn(a) { a; };
x(4);
`,
			expected: 4,
		},
		{
			input: `
let sum = fn(a, b) { a + b; };
sum(1, 2);
`,
			expected: 3,
		},
		{
			input: `
let sum = fn(a, b) {
	let c = a + b;
	c;
};
sum(1, 2);
`,
			expected: 3,
		},
		{
			input: `
let sum = fn(a, b) {
	let c = a + b;
	c;
};
sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			input: `
let sum = fn(a, b) {
	let c = a + b;
	c;
};
let outer = fn() {
	sum(1, 2) + sum(3, 4);
};
outer();
`,
			expected: 10,
		},
		{
			input: `
let globalNum = 10;

let sum = fn(a, b) {
	let c = a + b;
	c + globalNum;
};

let outer = fn() {
	sum(1, 2) + sum(3, 4) + globalNum;
};

outer() + globalNum;
`,
			expected: 50,
		},
		{
			input: `
let one = fn() { 1; };

let two = fn() { 
	let result = one(); 
	return result + result; 
};

let three = fn(two) { 
	two() + 1; 
};

three(two);
`,
			expected: 3,
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineCallingFunctionsWithBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
let x = fn() { let x = 1; x };
x();
`,
			expected: 1,
		},
		{
			input: `
let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
oneAndTwo();
`,
			expected: 3,
		},
		{
			input: `
let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
oneAndTwo() + threeAndFour();
`,
			expected: 10,
		},
		{
			input: `
let firstFoobar = fn() { let foobar = 50; foobar; };
let secondFoobar = fn() { let foobar = 100; foobar; };
firstFoobar() + secondFoobar();
`,
			expected: 150,
		},
		{
			input: `
let globalSeed = 50;
let minusOne = fn() {
	let num = 1;
	globalSeed - num;
}
let minusTwo = fn() {
	let num = 2;
	globalSeed - num;
}
minusOne() + minusTwo();
`,
			expected: 97,
		},
		{
			input: `
let returnsOneReturner = fn() {
let returnsOne = fn() { 1; };
returnsOne;
};
returnsOneReturner()();
`,
			expected: 1,
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineFunctionsWithoutReturnValue(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
let noReturn = fn() { };
noReturn();
`,
			expected: Null,
		},
		{
			input: `
let noReturnOne = fn() { };
let noReturnTwo = fn() { noReturnOne(); };
noReturnOne();
noReturnTwo();
`,
			expected: Null,
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineCallingFunctionsWithNoArguments(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
let func = fn() { 1 + 2; };
func();
`,
			expected: 3,
		},
		{
			input: `
let one = fn() { 1; };
let two = fn() { 2; };
one() + two()
`,
			expected: 3,
		},
		{
			input: `
let x = fn() { 1 };
let y = fn() { x() + 1 };
let z = fn() { y() + 1 };
z();
`,
			expected: 3,
		},
		{
			input: `
let earlyExit = fn() { return 1; 2; };
earlyExit();
`,
			expected: 1,
		},
		{
			input: `
let earlyExit = fn() { return 1; return 2; };
earlyExit();
`,
			expected: 1,
		},
		// First class function
		{
			input: `
let returnsOne = fn() { 1; };
let returnsOneReturner = fn() { returnsOne; };
returnsOneReturner()();
`,
			expected: 1,
		},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineIndexExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
	}
	runVirtualMachineTests(t, testCases)
}

func TestVirtualMachineMapLiterals(t *testing.T) {
	testCases := []vmTestCase{
		{
			"{}", map[object.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVirtualMachineTests(t, testCases)
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
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Map)
		if !ok {
			t.Errorf("actual not *object.Map: %T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("len(array.Items) = %d, want = %d", len(hash.Pairs), len(expected))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}
			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject() failed: %s", err)
			}
		}
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("actual not *object.Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("errObj.Message = %q, want = %q", errObj.Message, expected.Message)
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
