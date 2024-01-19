package ast

import "BigTalk_Interpreter/token"

type INode interface {
	TokenLiteral() string
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

func (l *LetStatement) statementNode() {}
func (l *LetStatement) TokenLiteral() string {
	return l.Token.Literal
}

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (l *Identifier) expressionNode() {}
func (l *Identifier) TokenLiteral() string {
	return l.Token.Literal
}

type ReturnStatement struct {
	Token token.Token // token.RETURN
	Value IExpression
}

func (r *ReturnStatement) statementNode() {

}
func (r *ReturnStatement) TokenLiteral() string {
	return r.Token.Literal
}
