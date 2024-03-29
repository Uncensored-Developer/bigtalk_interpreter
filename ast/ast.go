package ast

import (
	"BigTalk_Interpreter/token"
	"bytes"
	"fmt"
	"strings"
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

// LetStatement
// Basic Structure: let <identifier> = <expression>;
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

func (i *Identifier) expressionNode() {}
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
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

type PrefixExpression struct {
	Token    token.Token // prefix token i.e ! or -
	Operator string
	Value    IExpression
}

func (p *PrefixExpression) expressionNode() {

}
func (p *PrefixExpression) TokenLiteral() string {
	return p.Token.Literal
}

func (p *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(p.Operator)
	out.WriteString(p.Value.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token      token.Token // operator token, e.g +, *
	LeftValue  IExpression
	Operator   string
	RightValue IExpression
}

func (i *InfixExpression) expressionNode() {

}
func (i *InfixExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(i.LeftValue.String())
	out.WriteString(" " + i.Operator + " ")
	out.WriteString(i.RightValue.String())
	out.WriteString(")")

	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {

}

func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) String() string {
	return b.Token.Literal
}

type BlockStatement struct {
	Token      token.Token // token.LBRACE
	Statements []IStatement
}

func (b *BlockStatement) statementNode() {

}

func (b *BlockStatement) TokenLiteral() string {
	return b.Token.Literal
}

func (b *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range b.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type IfExpression struct {
	Token       token.Token // token.IF
	Condition   IExpression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfExpression) expressionNode() {

}

func (i *IfExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(i.Condition.String())
	out.WriteString(" ")
	out.WriteString(i.Consequence.String())

	if i.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(i.Alternative.String())
	}

	return out.String()
}

type FunctionLiteral struct {
	Name       string
	Token      token.Token // token.FUNCTION
	Parameters []*Identifier
	Body       *BlockStatement
}

func (f *FunctionLiteral) expressionNode() {

}

func (f *FunctionLiteral) TokenLiteral() string {
	return f.Token.Literal
}

func (f *FunctionLiteral) String() string {
	var out bytes.Buffer

	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(f.TokenLiteral())
	if f.Name != "" {
		out.WriteString(fmt.Sprintf("<%s>", f.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(f.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token // token.LPAREN
	Func      IExpression //  FunctionLiteral or Identifier
	Arguments []IExpression
}

func (c *CallExpression) expressionNode() {

}

func (c *CallExpression) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CallExpression) String() string {
	var out bytes.Buffer

	var args []string
	for _, a := range c.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(c.Func.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (s *StringLiteral) expressionNode() {

}

func (s *StringLiteral) TokenLiteral() string {
	return s.Token.Literal
}

func (s *StringLiteral) String() string {
	return s.Token.Literal
}

// ArrayLiteral
// Basic structure: [<expression>, <expression>, ...]
type ArrayLiteral struct {
	Token token.Token // token.L_SQR_BRACKET
	Items []IExpression
}

func (a *ArrayLiteral) expressionNode() {

}

func (a *ArrayLiteral) TokenLiteral() string {
	return a.Token.Literal
}

func (a *ArrayLiteral) String() string {
	var out bytes.Buffer

	var elements []string
	for _, item := range a.Items {
		elements = append(elements, item.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// IndexExpression
// Basic structure: <expression>[<expression>]
type IndexExpression struct {
	Token token.Token // token.L_SQR_BRACKET
	Left  IExpression
	Index IExpression
}

func (i *IndexExpression) expressionNode() {

}

func (i *IndexExpression) TokenLiteral() string {
	return i.Token.Literal
}

func (i *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(i.Left.String())
	out.WriteString("[")
	out.WriteString(i.Index.String())
	out.WriteString("])")

	return out.String()
}

// MapLiteral
// Basic structure: {<expression> : <expression>, <expression> : <expression>, ... }
type MapLiteral struct {
	Token token.Token // token.LBRACE
	Pairs map[IExpression]IExpression
}

func (m *MapLiteral) expressionNode() {

}

func (m *MapLiteral) TokenLiteral() string {
	return m.Token.Literal
}

func (m *MapLiteral) String() string {
	var out bytes.Buffer

	var pairs []string
	for key, value := range m.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
