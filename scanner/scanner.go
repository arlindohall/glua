package scanner

import (
	"arlindohall/glua/glerror"
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type Token struct {
	Text string
	Type TokenType
	Line int
}

type TokenType int

type scanner struct {
	reader *bufio.Reader
	line   int
	err    glerror.GluaErrorChain
}

const (
	TokenError = iota
	TokenAnd
	TokenAssert
	TokenBang
	TokenCaret
	TokenComma
	TokenDo
	TokenDot
	TokenElse
	TokenEnd
	TokenEof
	TokenEqual
	TokenEqualEqual
	TokenFalse
	TokenFor
	TokenFunction
	TokenGlobal
	TokenGreater
	TokenGreaterEqual
	TokenIdentifier
	TokenIf
	TokenIn
	TokenLeftBrace
	TokenLeftBracket
	TokenLeftParen
	TokenLess
	TokenLessEqual
	TokenLocal
	TokenMinus
	TokenNil
	TokenNumber
	TokenOr
	TokenPlus
	TokenReturn
	TokenRightBrace
	TokenRightBracket
	TokenRightParen
	TokenSemicolon
	TokenSlash
	TokenStar
	TokenString
	TokenThen
	TokenTildeEqual
	TokenTrue
	TokenWhile
)

func Scanner(reader *bufio.Reader) *scanner {
	return &scanner{
		reader: reader,
		err:    glerror.GluaErrorChain{},
		line:   1,
	}
}

// todo: goroutine publishes one token at a time?
func (scanner *scanner) ScanTokens() ([]Token, glerror.GluaErrorChain) {
	var tokens []Token

	token, err := scanner.scanToken()
	for ; err == nil; token, err = scanner.scanToken() {
		tokens = append(tokens, token)
	}

	if err == io.EOF {
		return tokens, scanner.err
	}

	return tokens, scanner.err
}

func (scanner *scanner) peekRune() (rune, error) {
	r, _, err := scanner.reader.ReadRune()

	if err == io.EOF {
		return 0, err
	}

	if err != nil {
		scanner.error(err.Error())
		return 0, err
	}

	err = scanner.reader.UnreadRune()

	if err != nil {
		scanner.error(err.Error())
		return 0, err
	}

	return r, nil
}

func (scanner *scanner) scanRune() (rune, error) {
	r, _, err := scanner.reader.ReadRune()

	if err == io.EOF {
		return 0, err
	}

	if err != nil {
		scanner.error(err.Error())
		return 0, err
	}

	return r, nil
}

func (scanner *scanner) advance() (ok bool) {
	_, err := scanner.scanRune()

	ok = err == io.EOF || err == nil
	return
}

func (scanner *scanner) revert() {
	err := scanner.reader.UnreadRune()

	if err != nil {
		scanner.error(err.Error())
	}
}

func (scanner *scanner) check(r rune) bool {
	next, err := scanner.peekRune()

	if err == nil && next == r {
		scanner.advance()
		return true
	} else {
		return false
	}
}

func (scanner *scanner) skipWhitespace() *Token {
	for r, err := scanner.peekRune(); err == nil; r, err = scanner.peekRune() {
		if r == '/' {
			if scanner.consumeComment() {
				return nil
			}
			token := scanner.makeToken("/", TokenSlash)
			return &token
		}

		if !unicode.IsSpace(r) {
			return nil
		}

		if r == '\n' {
			scanner.line += 1
		}

		scanner.advance()
	}

	return nil
}

func (scanner *scanner) consumeComment() bool {
	scanner.advance()
	secondSlash, _ := scanner.peekRune()

	if secondSlash != '/' {
		return false
	}

	var err error
	var r rune
	for r, err = scanner.peekRune(); err == nil && r != '\n'; r, err = scanner.peekRune() {
		scanner.advance()
	}

	if err == nil {
		scanner.advance()
	}

	return true
}

func (scanner *scanner) scanToken() (Token, error) {
	token := scanner.skipWhitespace()

	if token != nil {
		return *token, nil
	}

	r, err := scanner.peekRune()
	switch {
	case err == io.EOF:
		return scanner.makeToken("", TokenEof), err
	case err != nil:
		scanner.error(fmt.Sprint("Error reading next character ", err))
		return scanner.makeToken("", TokenError), nil
	case isNumber(r):
		return scanner.scanNumber()
	case isAlpha(r):
		return scanner.scanWord()
	case scanner.check('+'):
		return scanner.makeToken("+", TokenPlus), nil
	case scanner.check('-'):
		return scanner.makeToken("-", TokenMinus), nil
	case scanner.check('*'):
		return scanner.makeToken("*", TokenStar), nil
	case scanner.check('/'):
		return scanner.makeToken("/", TokenSlash), nil
	case scanner.check(';'):
		return scanner.makeToken(";", TokenSemicolon), nil
	case scanner.check('!'):
		return scanner.makeToken("!", TokenBang), nil
	case scanner.check('<'):
		if scanner.check('=') {
			return scanner.makeToken("<=", TokenLessEqual), nil
		} else {
			return scanner.makeToken("<", TokenLess), nil
		}
	case scanner.check('>'):
		if scanner.check('=') {
			return scanner.makeToken(">=", TokenGreaterEqual), nil
		} else {
			return scanner.makeToken(">", TokenGreater), nil
		}
	case scanner.check('='):
		if scanner.check('=') {
			return scanner.makeToken("==", TokenEqualEqual), nil
		} else {
			return scanner.makeToken("=", TokenEqual), nil
		}
	case scanner.check('"'):
		return scanner.scanString()
	case scanner.check('{'):
		return scanner.makeToken("{", TokenLeftBrace), nil
	case scanner.check('}'):
		return scanner.makeToken("}", TokenRightBrace), nil
	case scanner.check('['):
		return scanner.makeToken("[", TokenLeftBracket), nil
	case scanner.check(']'):
		return scanner.makeToken("]", TokenRightBracket), nil
	case scanner.check('('):
		return scanner.makeToken("(", TokenLeftParen), nil
	case scanner.check(')'):
		return scanner.makeToken(")", TokenRightParen), nil
	case scanner.check(','):
		return scanner.makeToken(",", TokenComma), nil
	case scanner.check('.'):
		return scanner.makeToken(".", TokenDot), nil
	default:
		scanner.advance()
		scanner.error(fmt.Sprint("Unexpected character '", string([]rune{r}), "'"))
		return scanner.makeToken("", TokenError), scanner.err
	}
}

func (scanner *scanner) makeToken(name string, tt TokenType) Token {
	return Token{
		Text: name,
		Type: tt,
		Line: scanner.line,
	}
}

func (scanner *scanner) scanNumber() (Token, error) {
	var runes []rune

	// todo: decimals
	r, err := scanner.scanRune()
	for ; err == nil && isNumber(r) || r == '_'; r, err = scanner.scanRune() {
		if r == '_' {
			continue
		}

		runes = append(runes, r)
	}

	// On EOF, just unread the EOF and let the next skipWhitespace pick it up
	if err != nil && err != io.EOF {
		return scanner.makeToken(string(runes), TokenError), err
	}

	if err != io.EOF {
		scanner.revert()
	}

	return scanner.makeToken(
		string(runes),
		TokenNumber,
	), nil
}

func (scanner *scanner) scanString() (Token, error) {
	var literal []rune
	for r, err := scanner.scanRune(); err == nil && r != '"'; r, err = scanner.scanRune() {
		if r == '\n' {
			scanner.line += 1
			scanner.error("Newline in string literal")
			return scanner.makeToken(string(literal), TokenError), scanner.err
		}

		if r != '\\' {
			literal = append(literal, r)
			continue
		}

		escape, escapeErr := scanner.scanEscape()

		if escapeErr != nil {
			return scanner.makeToken(string(literal), TokenError), escapeErr
		}

		literal = append(literal, escape)
	}

	return scanner.makeToken(string(literal), TokenString), nil
}

func (scanner *scanner) scanEscape() (rune, error) {
	r, err := scanner.scanRune()

	if err != nil {
		scanner.error("Failed to scan escape sequence")
		return 0, nil
	}

	switch r {
	case '\\':
		return '\\', nil
	case 'n':
		return '\n', nil
	case '"':
		return '"', nil
	default:
		scanner.error(fmt.Sprint("Invalid escape sequence: \\", r))
		return 0, scanner.err
	}
}

func (scanner *scanner) scanWord() (Token, error) {
	var word []rune
	for c, err := scanner.peekRune(); err == nil && isAlpha(c) || isNumber(c); c, err = scanner.peekRune() {
		word = append(word, c)
		if !scanner.advance() {
			break
		}
	}

	source := string(word)

	// todo: use a trie to speed up
	switch source {
	case "assert":
		return scanner.makeToken(source, TokenAssert), nil
	case "true":
		return scanner.makeToken(source, TokenTrue), nil
	case "false":
		return scanner.makeToken(source, TokenFalse), nil
	case "and":
		return scanner.makeToken(source, TokenAnd), nil
	case "or":
		return scanner.makeToken(source, TokenOr), nil
	case "nil":
		return scanner.makeToken(source, TokenNil), nil
	case "global":
		return scanner.makeToken(source, TokenGlobal), nil
	case "local":
		return scanner.makeToken(source, TokenLocal), nil
	case "while":
		return scanner.makeToken(source, TokenWhile), nil
	case "for":
		return scanner.makeToken(source, TokenFor), nil
	case "in":
		return scanner.makeToken(source, TokenIn), nil
	case "if":
		return scanner.makeToken(source, TokenIf), nil
	case "then":
		return scanner.makeToken(source, TokenThen), nil
	case "else":
		return scanner.makeToken(source, TokenElse), nil
	case "do":
		return scanner.makeToken(source, TokenDo), nil
	case "end":
		return scanner.makeToken(source, TokenEnd), nil
	case "function":
		return scanner.makeToken(source, TokenFunction), nil
	case "return":
		return scanner.makeToken(source, TokenReturn), nil
	default:
		return scanner.makeToken(source, TokenIdentifier), nil
	}

}

func isNumber(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func isAlpha(r rune) bool {
	lower := unicode.ToLower(r)
	return 'a' <= lower && 'z' >= lower
}

func (scanner *scanner) error(message string) {
	scanner.err.Append(ScanError{
		message: message,
		line:    scanner.line,
	})
}

type ScanError struct {
	message string
	line    int
}

// todo: line numbers
func (se ScanError) Error() string {
	return fmt.Sprintf("Scan error [line=%d] ---> %s", se.line, se.message)
}
