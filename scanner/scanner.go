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
}

type TokenType int

type scanner struct {
	reader *bufio.Reader
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
	TokenEnd
	TokenEof
	TokenEqual
	TokenEqualEqual
	TokenFalse
	TokenGlobal
	TokenGreater
	TokenGreaterEqual
	TokenIdentifier
	TokenLeftBrace
	TokenLeftBracket
	TokenLess
	TokenLessEqual
	TokenMinus
	TokenNil
	TokenNumber
	TokenOr
	TokenPlus
	TokenRightBrace
	TokenRightBracket
	TokenSemicolon
	TokenSlash
	TokenStar
	TokenString
	TokenTildeEqual
	TokenTrue
	TokenWhile
)

func Scanner(reader *bufio.Reader) *scanner {
	return &scanner{
		reader: reader,
		err:    glerror.GluaErrorChain{},
	}
}

func (scanner *scanner) ScanTokens() ([]Token, glerror.GluaErrorChain) {
	var tokens []Token

	token, err := scanner.scanToken()
	for ; err == nil; token, err = scanner.scanToken() {
		tokens = append(tokens, token)
	}

	if err == io.EOF {
		return tokens, scanner.err
	}

	scanner.error(fmt.Sprint("Error reading from file ", scanner.err))

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

func (scanner *scanner) skipWhitespace() {
	for r, err := scanner.peekRune(); err == nil; r, err = scanner.peekRune() {
		if !unicode.IsSpace(r) {
			return
		}

		scanner.advance()
	}
}

func (scanner *scanner) scanToken() (Token, error) {
	scanner.skipWhitespace()

	r, err := scanner.peekRune()
	switch {
	case err == io.EOF:
		return Token{"", TokenEof}, err
	case err != nil:
		scanner.error(fmt.Sprint("Error reading next character ", err))
		return Token{"", TokenError}, nil
	case isNumber(r):
		return scanner.scanNumber()
	case isAlpha(r):
		return scanner.scanWord()
	case scanner.check('+'):
		return Token{"+", TokenPlus}, nil
	case scanner.check('-'):
		return Token{"-", TokenMinus}, nil
	case scanner.check('*'):
		return Token{"*", TokenStar}, nil
	case scanner.check('/'):
		return Token{"/", TokenSlash}, nil
	case scanner.check(';'):
		return Token{";", TokenSemicolon}, nil
	case scanner.check('!'):
		return Token{"!", TokenBang}, nil
	case scanner.check('<'):
		if scanner.check('=') {
			return Token{"=", TokenEqual}, nil
		} else {
			return Token{"<", TokenLess}, nil
		}
	case scanner.check('='):
		if scanner.check('=') {
			return Token{"==", TokenEqualEqual}, nil
		} else {
			return Token{"=", TokenEqual}, nil
		}
	case scanner.check('"'):
		return scanner.scanString()
	case scanner.check('{'):
		return Token{"{", TokenLeftBrace}, nil
	case scanner.check('}'):
		return Token{"}", TokenRightBrace}, nil
	case scanner.check('['):
		return Token{"[", TokenLeftBracket}, nil
	case scanner.check(']'):
		return Token{"]", TokenRightBracket}, nil
	case scanner.check(','):
		return Token{",", TokenComma}, nil
	case scanner.check('.'):
		return Token{".", TokenDot}, nil
	default:
		scanner.advance()
		scanner.error(fmt.Sprint("Unexpected character '", string([]rune{r}), "'"))
		return Token{}, scanner.err
	}
}

func (scanner *scanner) scanNumber() (Token, error) {
	var runes []rune

	r, err := scanner.scanRune()
	for ; err == nil && isNumber(r); r, err = scanner.scanRune() {
		runes = append(runes, r)
	}

	// On EOF, just unread the EOF and let the next skipWhitespace pick it up
	if err != nil && err != io.EOF {
		return Token{}, err
	}

	if err != io.EOF {
		scanner.revert()
	}

	return Token{
		Text: string(runes),
		Type: TokenNumber,
	}, nil
}

func (scanner *scanner) scanString() (Token, error) {
	var literal []rune
	for r, err := scanner.scanRune(); err == nil && r != '"'; r, err = scanner.scanRune() {
		if r == '\n' {
			scanner.error("Newline in string literal")
			return Token{string(literal), TokenError}, scanner.err
		}

		if r != '\\' {
			literal = append(literal, r)
			continue
		}

		escape, escapeErr := scanner.scanEscape()

		if escapeErr != nil {
			return Token{"", TokenError}, escapeErr
		}

		literal = append(literal, escape)
	}

	return Token{string(literal), TokenString}, nil
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
		return Token{source, TokenAssert}, nil
	case "true":
		return Token{source, TokenTrue}, nil
	case "false":
		return Token{source, TokenFalse}, nil
	case "and":
		return Token{source, TokenAnd}, nil
	case "or":
		return Token{source, TokenOr}, nil
	case "nil":
		return Token{source, TokenNil}, nil
	case "global":
		return Token{source, TokenGlobal}, nil
	case "while":
		return Token{source, TokenWhile}, nil
	case "do":
		return Token{source, TokenDo}, nil
	case "end":
		return Token{source, TokenEnd}, nil
	default:
		return Token{source, TokenIdentifier}, nil
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
	scanner.err.Append(ScanError{message})
}

type ScanError struct {
	message string
}

// todo: line numbers
func (se ScanError) Error() string {
	return fmt.Sprintf("Scan error ---> %s", se.message)
}
