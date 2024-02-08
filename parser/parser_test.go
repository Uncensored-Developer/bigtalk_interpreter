package parser

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/lexer"
	"fmt"
	"testing"
)

func TestParsingArrayLiteral(t *testing.T) {
	input := "[1, 2 * 3, 4 + 5]"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Value.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt is not ast.ArrayLiteral, got = %T", stmt.Value)
	}

	if len(array.Items) != 3 {
		t.Fatalf("len(array.Items) = %d, want %d", len(array.Items), 3)
	}

	testIntegerLiteral(t, array.Items[0], 1)
	testInfixExpression(t, array.Items[1], 2, "*", 3)
	testInfixExpression(t, array.Items[2], 4, "+", 5)
}

func TestParsingStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("stmt.Value not *ast.StringLiteral, got = %T", stmt.Value)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value = %q, want = %q", literal.Value, "hello world")
	}
}

func TestParsingCallExpression(t *testing.T) {
	input := "add(1, 2 + 3, 4 * 5);"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	callExp, ok := stmt.Value.(*ast.CallExpression)
	if !ok {
		t.Fatalf("callExp not *ast.CallExpression. got=%T", stmt.Value)
	}

	if !testIdentifier(t, callExp.Func, "add") {
		return
	}

	if len(callExp.Arguments) != 3 {
		t.Fatalf("len(callExp.Arguments) = %d, want %d", len(callExp.Arguments), 3)
	}

	testLiteralExpression(t, callExp.Arguments[0], 1)
	testInfixExpression(t, callExp.Arguments[1], 2, "+", 3)
	testInfixExpression(t, callExp.Arguments[2], 4, "*", 5)
}

func TestParsingFunctionParameter(t *testing.T) {
	testCases := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		fnLit := stmt.Value.(*ast.FunctionLiteral)

		if len(fnLit.Parameters) != len(tc.expectedParams) {
			t.Errorf("len(fnLit.Parameters) = %d, want %d", len(fnLit.Parameters), len(tc.expectedParams))
		}

		for i, ident := range tc.expectedParams {
			testLiteralExpression(t, fnLit.Parameters[i], ident)
		}
	}
}

func TestParsingFunctionLiteral(t *testing.T) {
	input := "fn(x,y) { x + y; }"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	fnLit, ok := stmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("fnLit not *ast.FunctionLiteral. got=%T", stmt.Value)
	}

	if len(fnLit.Parameters) != 2 {
		t.Fatalf("len(fnLit.Parameters) = %d, want %d", len(fnLit.Parameters), 2)
	}

	testLiteralExpression(t, fnLit.Parameters[0], "x")
	testLiteralExpression(t, fnLit.Parameters[1], "y")

	if len(fnLit.Body.Statements) != 1 {
		t.Fatalf("len(fnLit.Body.Statements) = %d, want %d", len(fnLit.Body.Statements), 1)
	}

	bodyStmt, ok := fnLit.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("fnLit.Body.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Value, "x", "+", "y")
}

func TestParsingIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	ifExp, ok := stmt.Value.(*ast.IfExpression)
	if !ok {
		t.Fatalf("exp not *ast.IfExpression. got=%T", stmt.Value)
	}

	if !testInfixExpression(t, ifExp.Condition, "x", "<", "y") {
		return
	}

	if len(ifExp.Consequence.Statements) != 1 {
		t.Errorf("len(ifExp.Consequence.Statements) = %d, want %d", len(ifExp.Consequence.Statements), 1)
	}

	consequence, ok := ifExp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got = %T",
			ifExp.Consequence.Statements[0])
	}
	if !testIdentifier(t, consequence.Value, "x") {
		return
	}
	if ifExp.Alternative != nil {
		t.Errorf("ifExp.Alternative = %+v. want <nil>", ifExp.Alternative)
	}
}

func TestParsingIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"

	l := lexer.NewLexer(input)
	p := NewParser(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
			program.Statements[0])
	}

	ifExp, ok := stmt.Value.(*ast.IfExpression)
	if !ok {
		t.Fatalf("exp not *ast.IfExpression. got=%T", stmt.Value)
	}

	if !testInfixExpression(t, ifExp.Condition, "x", "<", "y") {
		return
	}

	if len(ifExp.Consequence.Statements) != 1 {
		t.Errorf("len(ifExp.Consequence.Statements) = %d, want %d", len(ifExp.Consequence.Statements), 1)
	}

	consequence, ok := ifExp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got = %T",
			ifExp.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Value, "x") {
		return
	}

	if len(ifExp.Alternative.Statements) != 1 {
		t.Errorf("len(ifExp.Alternative.Statements) = %d, want %d", len(ifExp.Alternative.Statements), 1)
	}

	alternative, ok := ifExp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got = %T",
			ifExp.Consequence.Statements[0])
	}

	if !testIdentifier(t, alternative.Value, "y") {
		return
	}
}

func TestParsingBooleanExpression(t *testing.T) {
	testCases := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got = %T",
				program.Statements[0])
		}

		boolExp, ok := stmt.Value.(*ast.Boolean)
		if !ok {
			t.Fatalf("exp not *ast.Boolean. got=%T", stmt.Value)
		}
		if boolExp.Value != tc.expectedBoolean {
			t.Errorf("boolExp.Value = %t, want %t", boolExp.Value, tc.expectedBoolean)
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},

		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},

		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		got := program.String()
		if got != tc.expected {
			t.Errorf("program.String() = %q, expected %q", got, tc.expected)
		}
	}
}

func TestParsingInfixExpression(t *testing.T) {
	testCases := []struct {
		input      string
		leftValue  any
		operator   string
		rightValue any
	}{
		{"5 + 3;", 5, "+", 3},
		{"5 * 3;", 5, "*", 3},
		{"5 - 3;", 5, "-", 3},
		{"5 > 6;", 5, ">", 6},
		{"5 < 3;", 5, "<", 3},
		{"5 / 3;", 5, "/", 3},
		{"5 == 3;", 5, "==", 3},
		{"5 != 3;", 5, "!=", 3},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
		{"foobar + barfoo;", "foobar", "+", "barfoo"},
		{"foobar - barfoo;", "foobar", "-", "barfoo"},
		{"foobar * barfoo;", "foobar", "*", "barfoo"},
		{"foobar / barfoo;", "foobar", "/", "barfoo"},
		{"foobar > barfoo;", "foobar", ">", "barfoo"},
		{"foobar < barfoo;", "foobar", "<", "barfoo"},
		{"foobar == barfoo;", "foobar", "==", "barfoo"},
		{"foobar != barfoo;", "foobar", "!=", "barfoo"},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got %T", program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Value, tc.leftValue, tc.operator, tc.rightValue) {
			return
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		operator string
		value    any
	}{
		{"-3", "-", 3},
		{"!13", "!", 13},
		{"!true;", "!", true},
		{"!false;", "!", false},
		{"-foobar;", "-", "foobar"},
		{"!foobar;", "!", "foobar"},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got %T", program.Statements[0])
		}

		prefixExp, ok := stmt.Value.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. got %T", stmt.Value)
		}

		if prefixExp.Operator != tc.operator {
			t.Fatalf("exp.Operator = %q, want %q", prefixExp.Operator, tc.operator)
		}

		if !testLiteralExpression(t, prefixExp.Value, tc.value) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, intLit ast.IExpression, value int64) bool {
	intgr, ok := intLit.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("intLit is not *ast.IntegerLiteral, got %T", intLit)
		return false
	}

	if intgr.Value != value {
		t.Errorf("intgr.Value = %d, want %d", intgr.Value, value)
		return false
	}

	if intgr.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("intgr.TokenLiteral() = %s, want %d", intgr.TokenLiteral(), value)
		return false
	}
	return true
}

func TestParsingIntegerLiteralExpression(t *testing.T) {
	input := "3;"

	l := lexer.NewLexer(input)
	p := NewParser(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got %T", program.Statements[0])
	}

	literal, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got %T", stmt.Value)
	}

	if literal.Value != 3 {
		t.Errorf("ident.Value = %q, want %q", literal.Value, "foo")
	}

	if literal.TokenLiteral() != "3" {
		t.Errorf("ident.TokenLiteral = %q, want %q", literal.TokenLiteral(), "foo")
	}
}

func TestParsingIdentifierExpression(t *testing.T) {
	input := "foo;"

	l := lexer.NewLexer(input)
	p := NewParser(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("len(program.Statements) = %d, want %d", len(program.Statements), 1)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got %T", program.Statements[0])
	}

	ident, ok := stmt.Value.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got %T", stmt.Value)
	}

	if ident.Value != "foo" {
		t.Errorf("ident.Value = %q, want %q", ident.Value, "foo")
	}

	if ident.TokenLiteral() != "foo" {
		t.Errorf("ident.TokenLiteral = %q, want %q", ident.TokenLiteral(), "foo")
	}
}

func TestParsingReturnStatements(t *testing.T) {
	testCases := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 3;", 3},
		{"return true;", true},
		{"return foo;", "foo"},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("len(program.Statements) = %d, want %d",
				len(program.Statements), 1)
		}

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.returnStatement. got = %T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Fatalf("returnStmt.TokenLiteral() = %q, want return",
				returnStmt.TokenLiteral())
		}
		if testLiteralExpression(t, returnStmt.Value, tc.expectedValue) {
			return
		}
	}
}

func TestParsingLetStatements(t *testing.T) {
	testCases := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foo = y;", "foo", "y"},
	}

	for _, tc := range testCases {
		l := lexer.NewLexer(tc.input)
		p := NewParser(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("len(program.Statements) = %d, want %d",
				len(program.Statements), 1)
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tc.expectedIdentifier) {
			return
		}
		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tc.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.IStatement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let', got %q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement, got %T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s', got %s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s', got %s", name, letStmt.Name)
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))

	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testIdentifier(t *testing.T, exp ast.IExpression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp = %T, want *ast.Identifier", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value = %s, want %s", ident.Value, value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral() = %s, want %s", ident.TokenLiteral(), value)
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, exp ast.IExpression, expected any) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	default:
		t.Errorf("exp type not handled, got %T", exp)
		return false
	}
}

func testInfixExpression(t *testing.T, exp ast.IExpression, left any, operator string, right any) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp = %T, want *ast.OperatorExpression", exp)
		return false
	}

	if !testLiteralExpression(t, opExp.LeftValue, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("opExp.Operator = %q, want %q", opExp.Operator, operator)
		return false
	}

	if !testLiteralExpression(t, opExp.RightValue, right) {
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.IExpression, value bool) bool {
	boolExp, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp = %T, want *ast.Boolean", exp)
		return false
	}

	if boolExp.Value != value {
		t.Errorf("boolExp.Value = %t, want %t", boolExp.Value, value)
		return false
	}

	if boolExp.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("boolExp.TokenLiteral() = %s, want %t", boolExp.TokenLiteral(), value)
	}
	return true
}
