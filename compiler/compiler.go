package compiler

import (
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
	"strconv"
)

const (
	OpAdd = iota
	OpAssert
	OpConstant
	OpDivide
	OpNil
	OpMult
	OpNegate
	OpNot
	OpPop
	OpReturn
	OpSubtract
)

const (
	RunFileMode = iota
	ReplMode
)

type Op byte

type ReturnMode int

type compiler struct {
	text  []scanner.Token
	curr  int
	chunk Chunk
	err   error
	mode  ReturnMode
}

type Chunk struct {
	Bytecode  []Op
	Constants []value.Value
}

type Function struct {
	Chunk Chunk
	Name  string
}

func Compile(text []scanner.Token, mode ReturnMode) (Function, error) {
	compiler := compiler{
		text,
		0,
		Chunk{},
		nil,
		mode,
	}

	compiler.compile()

	return compiler.end()
}

func (comp *compiler) compile() {
	for comp.current().Type != scanner.TokenEof {
		comp.declaration()
	}
}

func (comp *compiler) current() scanner.Token {
	if comp.curr >= len(comp.text) {
		return scanner.Token{
			Type: scanner.TokenEof,
			Text: "",
		}
	}

	return comp.text[comp.curr]
}

// func (comp *compiler) peek() scanner.Token {
// 	if comp.curr >= len(comp.text)-1 {
// 		return scanner.Token{
// 			"",
// 			scanner.TokenEof,
// 		}
// 	}

// 	return comp.text[comp.curr]
// }

func (comp *compiler) declaration() {
	comp.statement()

	// Lua allows semicolons but they are not required
	if comp.current().Type == scanner.TokenSemicolon {
		comp.consume(scanner.TokenSemicolon)
	}
}

func (comp *compiler) statement() {
	switch comp.current().Type {
	case scanner.TokenAssert:
		comp.advance()
		comp.expression()
		comp.emitByte(OpAssert)
	default:
		comp.expression()
		comp.emitByte(OpPop)
	}
}

func (comp *compiler) expression() {
	comp.logicOr()
}

func (comp *compiler) logicOr() {
	comp.logicAnd()

	for {
		if comp.current().Type == scanner.TokenOr {
			comp.advance()
			comp.logicAnd()
			panic("todo logic or")
		} else {
			return
		}
	}
}

func (comp *compiler) logicAnd() {
	comp.comparison()

	for {
		if comp.current().Type == scanner.TokenAnd {
			comp.advance()
			comp.comparison()
			panic("todo logic and")
		} else {
			return
		}
	}
}

func (comp *compiler) comparison() {
	comp.term()

	for {
		switch comp.current().Type {
		case scanner.TokenLess, scanner.TokenGreater, scanner.TokenLessEqual, scanner.TokenGreaterEqual, scanner.TokenTildeEqual, scanner.TokenEqualEqual:
			comp.advance()
			comp.term()
			panic("todo logic compare")
		default:
			return
		}
	}
}

func (comp *compiler) term() {
	comp.factor()

	for {
		switch comp.current().Type {
		case scanner.TokenPlus:
			comp.advance()
			comp.factor()
			comp.emitByte(OpAdd)
		case scanner.TokenMinus:
			comp.advance()
			comp.factor()
			comp.emitByte(OpSubtract)
		default:
			return
		}
	}
}

func (comp *compiler) factor() {
	comp.unary()

	for {
		switch comp.current().Type {
		case scanner.TokenStar:
			comp.advance()
			comp.unary()
			comp.emitByte(OpMult)
		case scanner.TokenSlash:
			comp.advance()
			comp.unary()
			comp.emitByte(OpDivide)
		default:
			return
		}
	}

}

func (comp *compiler) unary() {
	switch comp.current().Type {
	case scanner.TokenMinus:
		comp.advance()
		comp.unary()
		comp.emitByte(OpNegate)
	case scanner.TokenBang:
		comp.advance()
		comp.unary()
		comp.emitByte(OpNot)
	default:
		comp.exponent()
		return
	}
}

func (comp *compiler) exponent() {
	comp.primary()

	if comp.current().Type == scanner.TokenCaret {
		comp.advance()
		comp.primary()
		panic("todo exponentiation")
	}
}

func (comp *compiler) primary() {
	switch comp.current().Type {
	case scanner.TokenTrue:
		b := comp.makeConstant(&value.Boolean{Val: true})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	case scanner.TokenFalse:
		b := comp.makeConstant(&value.Boolean{Val: false})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	case scanner.TokenNumber:
		flt, err := strconv.ParseFloat(
			comp.current().Text,
			64,
		)

		if err != nil {
			comp.error(fmt.Sprint("Cannot parse number: ", comp.current().Text))
		}

		b := comp.makeConstant(&value.Number{Val: flt})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	default:
		comp.error(fmt.Sprint("Unexpected token:", comp.current()))
		comp.advance()
	}
}

func (comp *compiler) advance() {
	comp.curr += 1
}

func (comp *compiler) consume(tt scanner.TokenType) {
	if comp.current().Type != tt {
		comp.error(fmt.Sprint("Expected type ", tt, ", found ", comp.current().Type))
	}

	comp.advance()
}

func (comp *compiler) makeConstant(value value.Value) byte {
	index := len(comp.chunk.Constants)

	// todo: error if too many constants
	comp.chunk.Constants = append(comp.chunk.Constants, value)

	return byte(index)
}

func (comp *compiler) emitBytes(b1, b2 byte) {
	comp.chunk.Bytecode = append(comp.chunk.Bytecode, Op(b1))
	comp.chunk.Bytecode = append(comp.chunk.Bytecode, Op(b2))
}

func (comp *compiler) emitByte(b byte) {
	comp.chunk.Bytecode = append(comp.chunk.Bytecode, Op(b))
}

func (comp *compiler) emitReturn() {
	last := len(comp.chunk.Bytecode) - 1
	if comp.mode == ReplMode && comp.chunk.Bytecode[last] == OpPop {
		comp.chunk.Bytecode[last] = OpReturn
	} else {
		comp.emitBytes(OpNil, OpReturn)
	}
}

func (comp *compiler) end() (Function, error) {
	comp.emitReturn()

	if comp.err != nil {
		return Function{}, comp.err
	}

	return Function{
		comp.chunk,
		"",
	}, nil
}

func (comp *compiler) error(message string) {
	comp.err = CompileError{message}
}

type CompileError struct {
	message string
}

// todo: track line numbers in tokens and print error line
func (ce CompileError) Error() string {
	return fmt.Sprintf("Compile error ---> %s", ce.message)
}
