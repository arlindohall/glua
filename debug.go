package main

import "fmt"

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

func (op op) String() string {
	switch op {
	case OpNil:
		return "OpNil"
	case OpReturn:
		return "OpReturn"
	default:
		panic("Unrecognized Op")
	}
}
