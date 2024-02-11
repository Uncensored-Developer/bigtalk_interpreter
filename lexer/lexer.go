package lexer

import (
	"BigTalk_Interpreter/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	chr          byte // current char under examination
}

func NewLexer(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

// readChar reads the next character from the input string and updates the lexer's state.
// If the read position is at the end of the input string, the current character is set to 0 to indicate the end of the input.
// Otherwise, the current character is set to the character at the read position in the input string.
// The position, readPosition, and readPosition fields are updated accordingly.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.chr = 0
	} else {
		l.chr = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// NextToken gets the next token from the input string and returns it. It uses a lexer to identify the type of token and its literal value.
// The function starts by eating any leading whitespace characters by calling `l.eatWhitespace()`.
// Then, it uses a switch statement to check the current character `l.chr` and assign the appropriate token type and literal value to the `tok` variable. The possible cases are: '=',
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.eatWhitespace()

	switch l.chr {
	case '=':
		if l.peekChar() == '=' {
			chr := l.chr
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(chr) + string(l.chr)}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.chr)}
		}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.chr)}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.chr)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.chr)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.chr)}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.chr)}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.chr)}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.chr)}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.chr)}
	case '!':
		if l.peekChar() == '=' {
			chr := l.chr
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(chr) + string(l.chr)}
		} else {
			tok = token.Token{Type: token.BANG, Literal: string(l.chr)}
		}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.chr)}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.chr)}
	case '<':
		tok = token.Token{Type: token.LT, Literal: string(l.chr)}
	case '>':
		tok = token.Token{Type: token.GT, Literal: string(l.chr)}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '[':
		tok = token.Token{Type: token.L_SQR_BRACKET, Literal: string(l.chr)}
	case ']':
		tok = token.Token{Type: token.R_SQR_BRACKET, Literal: string(l.chr)}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: string(l.chr)}
	default:
		if isLetter(l.chr) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		} else if isDigit(l.chr) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.chr)}
		}
	}
	l.readChar()
	return tok
}

// readIdentifier reads characters from the input string until a non-letter character is encountered.
// It returns the substring of the input string starting from the original position
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.chr) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// eatWhitespace iterates over the input string and moves the read position until it encounters a non-whitespace character.
// The whitespace characters it removes are ' ', '\t', '\n', and '\r'.
func (l *Lexer) eatWhitespace() {
	for l.chr == ' ' || l.chr == '\t' || l.chr == '\n' || l.chr == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.chr) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// peekChar returns the next character in the lexer's input string without advancing the read position.
// If the read position is at the end of the input string, it returns 0 to indicate the end of the input.
// Otherwise, it returns the character at the read position in the input string.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// readString reads a string from the input starting from the current position.
// It advances the lexer until it encounters a closing double quote character or reaches the end of the input.
// The substring between the current position and the position before the closing double quote is returned as the string.
// If the lexer reaches the end of the input without encountering a closing double quote, the substring until the end of the input is returned.
// The position of the lexer is updated accordingly.
func (l *Lexer) readString() string {
	pos := l.position + 1
	for {
		l.readChar()
		if l.chr == '"' || l.chr == 0 {
			break
		}
	}
	return l.input[pos:l.position]
}

func isLetter(chr byte) bool {
	return 'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z' || chr == '_'
}

func isDigit(chr byte) bool {
	return '0' <= chr && chr <= '9'
}
