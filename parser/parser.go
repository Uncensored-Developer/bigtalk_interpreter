package parser

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX // -x or !x
	CALL
	INDEX // aFunction(x)
)

type (
	prefixParseFn func() ast.IExpression
	infixParseFn  func(ast.IExpression) ast.IExpression
)

var precedences = map[token.TokenType]int{
	token.EQ:            EQUALS,
	token.NOT_EQ:        EQUALS,
	token.LT:            LESSGREATER,
	token.GT:            LESSGREATER,
	token.PLUS:          SUM,
	token.MINUS:         SUM,
	token.SLASH:         PRODUCT,
	token.ASTERISK:      PRODUCT,
	token.LPAREN:        CALL,
	token.L_SQR_BRACKET: INDEX,
}

type Parser struct {
	lexer  *lexer.Lexer
	errors []string

	currentToken token.Token
	peekToken    token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.L_SQR_BRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.L_SQR_BRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function registered for %s token", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.IStatement{}
	for p.currentToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.IStatement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currentToken}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseLetStatement parses a Let statement and returns a pointer to ast.LetStatement.
//
// p *Parser - A pointer to the Parser object.
//
// Returns:
// *ast.LetStatement - A pointer to the Let statement AST node.
//
// This function creates a new LetStatement node with the current token. It then expects the next token to be an IDENT token,
// returns nil if it is not and sets the Name field of the LetStatement to an Identifier node with the value of the current token.
// It then expects the next token to be an ASSIGN token, returns nil if it is not and advances the parser to the next token.
// It then calls parseExpression with a precedence of LOWEST to parse the value expression of the Let statement and assigns the result
// to the Value field of the LetStatement node.
// Finally, it consumes any optional SEMICOLON tokens and returns the LetStatement node.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currentToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if fn, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fn.Name = stmt.Name.Value
	}

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the given type and advances the parser
// to the next token if it is. It returns true if the next token is of the given
// type, otherwise it returns false.
//
// t token.TokenType
// bool
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currentToken}
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.IExpression {
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.IExpression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.IExpression {
	literal := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as a 64bit integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	literal.Value = value
	return literal
}

func (p *Parser) parsePrefixExpression() ast.IExpression {
	exp := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.nextToken()
	exp.Value = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(leftValue ast.IExpression) ast.IExpression {
	exp := &ast.InfixExpression{
		Token:     p.currentToken,
		Operator:  p.currentToken.Literal,
		LeftValue: leftValue,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	exp.RightValue = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseBoolean() ast.IExpression {
	return &ast.Boolean{Token: p.currentToken, Value: p.currentTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.IExpression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// parseIfExpression parses an if expression in the input and returns an ast.IExpression representing the parsed expression.
// It creates an ast.IfExpression object with the current token as the token attribute.
// If the next token is not token.LPAREN, it returns nil.
// It advances the parser to the next token.
// It calls parseExpression with the LOWEST precedence level to parse the condition expression and assigns it to the Condition attribute of the IfExpression object.
// If the next token is not token.RPAREN, it returns nil.
// If the next token is not token.LBRACE, it returns nil.
// It calls parseBlockStatement to parse the block statement and assigns it to the Consequence attribute of the IfExpression object.
// If the next token is token.ELSE, it advances the parser to the next token.
// If the next token is not token.LBRACE, it returns nil.
// It calls parseBlockStatement to parse the alternative block statement and assigns it to the Alternative attribute of the IfExpression object.
// It returns the IfExpression object.
// ast.IExpression is the interface implemented by all types of expressions in the abstract syntax tree.
func (p *Parser) parseIfExpression() ast.IExpression {
	expression := &ast.IfExpression{Token: p.currentToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currentToken}
	block.Statements = []ast.IStatement{}

	p.nextToken()

	for !p.currentTokenIs(token.RBRACE) && !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// parseFunctionLiteral parses a function literal and returns an expression of type ast.IExpression.
// It creates a new ast.FunctionLiteral with the current token as the Token field.
// It expects the next token to be of type token.LPAREN and advances the parser to the next token if it is.
// If the next token is not token.LPAREN, it returns nil.
// It then calls parseFunctionParameters to parse the function parameters and assigns the result to the Parameters field of fnLit.
// It expects the next token to be of type token.LBRACE and advances the parser to the next token if it is.
// If the next token is not token.LBRACE, it returns nil.
// Finally, it calls parseBlockStatement to parse the function body and assigns the result to the Body field of fnLit.
// It returns fnLit.
func (p *Parser) parseFunctionLiteral() ast.IExpression {
	fnLit := &ast.FunctionLiteral{Token: p.currentToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fnLit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fnLit.Body = p.parseBlockStatement()
	return fnLit
}

// parseFunctionParameters parses the function parameters and returns a slice of ast.Identifier pointers.
// If there are no parameters, it returns an empty slice.
//
// It first checks if the next token is `)` and advances the parser if true.
// Then it advances the parser to the next token.
// It creates a new ast.Identifier with the current token's information and appends it to the identifiers slice.
//
// The function continues the loop as long as the next token is `,`.
// Inside the loop, it advances the parser twice to skip the comma and the next token.
// It creates a new ast.Identifier with the current token's information and appends it to the identifiers slice.
//
// Finally, it checks if the next token is `)` using the expectPeek method.
// If it is not `)`, it returns nil. Otherwise, it returns the identifiers slice.
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		ident := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parseCallExpression(function ast.IExpression) ast.IExpression {
	exp := &ast.CallExpression{Token: p.currentToken, Func: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseCallArguments parses the arguments of a function call.
// It returns a slice of ast.IExpression representing the parsed arguments.
//
// If the next token is a right parenthesis, it advances the parser to the next token
// and returns an empty slice.
//
// If the next token is not a right parenthesis, it advances the parser to the next token
// and parses the next expression, appending it to the args slice.
// It continues parsing expressions separated by commas until no more commas are found.
//
// It then expects the next token to be a right parenthesis and if not, it returns nil.
//
// Returns:
// []ast.IExpression - a slice of ast.IExpression representing the parsed arguments or nil if an error occurred.
func (p *Parser) parseCallArguments() []ast.IExpression {
	var args []ast.IExpression

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

func (p *Parser) parseStringLiteral() ast.IExpression {
	return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.IExpression {
	array := &ast.ArrayLiteral{Token: p.currentToken}
	array.Items = p.parseExpressionList(token.R_SQR_BRACKET)
	return array
}

// parseExpressionList parses a comma-separated list of expressions until it encounters the end token type.
// It returns a list of ast.IExpression.
//
// end token.TokenType: The token type that marks the end of the expression list.
// []ast.IExpression: The list of parsed expressions.
//
// Example usage:
//
//	list := parseExpressionList(token.COMMA)
//	for _, expr := range list {
//	    fmt.Println(expr)
//	}
func (p *Parser) parseExpressionList(end token.TokenType) []ast.IExpression {
	var list []ast.IExpression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseIndexExpression(left ast.IExpression) ast.IExpression {
	exp := &ast.IndexExpression{Token: p.currentToken, Left: left}
	p.nextToken()

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.R_SQR_BRACKET) {
		return nil
	}

	return exp
}

// parseMapLiteral parses a map literal expression.
//
// It initializes a new MapLiteral object with the current token and an empty map.
//
// It then iterates over tokens until it encounters a right brace token (}).
// Inside the loop, it calls 'parseExpression' to parse the key expression, expects a colon token (:) using the 'expectPeek' method,
// and calls 'parseExpression' again to parse the value expression.
//
// It adds the key-value pair to the MapLiteral object's Pairs map.
// If a comma token (,) is not followed by a right brace token (}), it expects the next token to be a comma.
//
// Finally, it expects a right brace token (}) and returns the MapLiteral object.
//
// Returns: an instance of ast.IExpression representing a map literal.
func (p *Parser) parseMapLiteral() ast.IExpression {
	mapLit := &ast.MapLiteral{Token: p.currentToken}
	mapLit.Pairs = make(map[ast.IExpression]ast.IExpression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()

		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		mapLit.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return mapLit
}
