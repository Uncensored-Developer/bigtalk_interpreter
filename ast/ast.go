package ast

import (
	"BigTalk_Interpreter/token"
	"bytes"
)

type INode interface {
	TokenLiteral() string
	String() string
}

type IStatement interface {
	INode
	statementNode()
}

type IExpression interface {
	INode
	expressionNode()
}

type Program struct {
	Statements []IStatement
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value IExpression
}

func (l *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(l.TokenLiteral() + " ")
	out.WriteString(l.Name.String())
	out.WriteString(" = ")

	if l.Value != nil {
		out.WriteString(l.Value.String())
	}

	out.WriteString(";")
	return out.String()
}

func (l *LetStatement) statementNode() {}
func (l *LetStatement) TokenLiteral() string {
	return l.Token.Literal
}

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) String() string {
	return i.Value
}

func (l *Identifier) expressionNode() {}
func (l *Identifier) TokenLiteral() string {
	return l.Token.Literal
}

type ReturnStatement struct {
	Token token.Token // token.RETURN
	Value IExpression
}

func (r *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(r.TokenLiteral() + " ")

	if r.Value != nil {
		out.WriteString(r.Value.String())
	}

	out.WriteString(";")
	return out.String()
}

func (r *ReturnStatement) statementNode() {

}
func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Literal
}

type ExpressionStatement struct {
	Token token.Token
	Value IExpression
}

func (e *ExpressionStatement) String() string {
	if e.Value != nil {
		return e.Value.String()
	}
	return ""
}

func (e *ExpressionStatement) statementNode() {

}
func (e *ExpressionStatement) TokenLiteral() string {
	return e.Token.Literal
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (i *IntegerLiteral) expressionNode() {

}
func (i *IntegerLiteral) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IntegerLiteral) String() string {
	return i.Token.Literal
}
