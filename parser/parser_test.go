package parser

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/lexer"
	"testing"
)

func TestIntegerLiteralExpression(t *testing.T) {
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
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	literal, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Value)
	}

	if literal.Value != 3 {
		t.Errorf("ident.Value = %q, want %q", literal.Value, "foo")
	}

	if literal.TokenLiteral() != "3" {
		t.Errorf("ident.TokenLiteral = %q, want %q", literal.TokenLiteral(), "foo")
	}
}

func TestIdentifierExpression(t *testing.T) {
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
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Value.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Value)
	}

	if ident.Value != "foo" {
		t.Errorf("ident.Value = %q, want %q", ident.Value, "foo")
	}

	if ident.TokenLiteral() != "foo" {
		t.Errorf("ident.TokenLiteral = %q, want %q", ident.TokenLiteral(), "foo")
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
return 3;
return 12;
return 99999;
`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements, got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnStatment, got=%T", stmt)
			continue
		}

		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral() not 'return', got %q", returnStmt.TokenLiteral())
		}
	}
}

func TestLetStatements(t *testing.T) {
	input := `
let x = 3;
let y = 2;
let foo = 99999;
`

	l := lexer.NewLexer(input)
	p := NewParser(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements, got=%d", len(program.Statements))
	}

	testCases := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foo"},
	}

	for i, tc := range testCases {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tc.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.IStatement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let', got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement, got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s', got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s', got=%s", name, letStmt.Name)
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
