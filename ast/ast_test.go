package ast

import (
	"BigTalk_Interpreter/token"
	"testing"
)

func TestProgram_String(t *testing.T) {
	program := &Program{
		Statements: []IStatement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "varName"},
					Value: "varName",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "newVar"},
					Value: "newVar",
				},
			},
		},
	}
	want := "let varName = newVar;"
	got := program.String()
	if got != want {
		t.Errorf("program.String() = %q, want = %q", got, want)
	}
}
