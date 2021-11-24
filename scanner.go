package main

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type Token struct {
	text  string
	_type TokenType
}

type TokenType int

type scanner struct {
	reader *bufio.Reader
	err    error
}

const (
	TokenNumber = iota
	TokenEof
	TokenPlus
	TokenSemicolon
)

func Scanner(reader *bufio.Reader) *scanner {
	return &scanner{
		reader: reader,
		err:    nil,
	}
}

func (tt TokenType) String() string {
	switch tt {
	case TokenNumber:
		return "TokenNumber"
	case TokenEof:
		return "TokenEof"
	case TokenPlus:
		return "TokenPlus"
	case TokenSemicolon:
		return "TokenSemicolon"
	default:
		panic("Unrecognized TokenType")
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v/\"%v\"", t._type, t.text)
}

func (scanner *scanner) ScanTokens() ([]Token, error) {
	var tokens []Token

	token, err := scanner.scanToken()
	for ; err == nil; token, err = scanner.scanToken() {
		tokens = append(tokens, token)
	}

	if err == io.EOF {
		return tokens, nil
	}

	fmt.Println("Error reading from file", scanner.err)

	return tokens, err
}

func (scanner *scanner) peekRune() (rune, error) {
	r, _, err := scanner.reader.ReadRune()

	if err != nil {
		scanner.err = err
		return 0, err
	}

	err = scanner.reader.UnreadRune()

	if err != nil {
		scanner.err = err
		return 0, err
	}

	return r, nil
}

func (scanner *scanner) scanRune() (rune, error) {
	r, _, err := scanner.reader.ReadRune()

	if err != nil {
		scanner.err = err
		return 0, err
	}

	return r, nil
}

func (scanner *scanner) advance() {
	_, _, err := scanner.reader.ReadRune()

	if err != nil {
		panic(err)
	}
}

func (scanner *scanner) revert() {
	err := scanner.reader.UnreadRune()

	if err != nil {
		panic(err)
	}
}

func (scanner *scanner) skipWhitespace() {
	for r, _, err := scanner.reader.ReadRune(); err == nil; r, _, err = scanner.reader.ReadRune() {
		if unicode.IsSpace(r) {
			continue
		}

		scanner.revert()
		return
	}
}

func (scanner *scanner) scanToken() (Token, error) {
	scanner.skipWhitespace()

	r, err := scanner.peekRune()
	switch {
	case err == io.EOF:
		return Token{"", TokenEof}, err
	case err != nil:
		panic(fmt.Sprint("Error reading next character ", err))
	case isNumber(r):
		return scanner.scanNumber()
	case isAlpha(r):
		return scanner.scanWord()
	case r == '+':
		scanner.advance()
		return Token{"+", TokenPlus}, nil
	case r == ';':
		scanner.advance()
		return Token{";", TokenSemicolon}, nil
	default:
		panic(fmt.Sprint("Unexpected character '", string([]rune{r}), "'"))
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

	scanner.revert()

	return Token{
		text:  string(runes),
		_type: TokenNumber,
	}, nil
}

func (scanner *scanner) scanWord() (Token, error) {
	return Token{}, nil
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
	switch unicode.ToLower(r) {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
		'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		return true
	default:
		return false
	}
}
