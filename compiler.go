package main

import (
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
	OpPop
	OpReturn
	OpSubtract
)

type compiler struct {
	text  []Token
	curr  int
	chunk chunk
	err   error
	mode  ReturnMode
}

type chunk struct {
	bytecode  []op
	constants []Value
}

type Function struct {
	chunk chunk
	name  string
}

func Compile(text []Token, mode ReturnMode) (Function, error) {
	compiler := compiler{
		text,
		0,
		chunk{},
		nil,
		mode,
	}

	compiler.compile()

	return compiler.end()
}

func (comp *compiler) compile() {
	for comp.current()._type != TokenEof {
		comp.declaration()
	}
}

func (comp *compiler) current() Token {
	if comp.curr >= len(comp.text) {
		return Token{
			"",
			TokenEof,
		}
	}

	return comp.text[comp.curr]
}

// func (comp *compiler) peek() Token {
// 	if comp.curr >= len(comp.text)-1 {
// 		return Token{
// 			"",
// 			TokenEof,
// 		}
// 	}

// 	return comp.text[comp.curr]
// }

func (comp *compiler) declaration() {
	comp.statement()

	// Lua allows semicolons but they are not required
	if comp.current()._type == TokenSemicolon {
		comp.consume(TokenSemicolon)
	}
}

func (comp *compiler) statement() {
	switch comp.current()._type {
	case TokenAssert:
		comp.advance()
		comp.expression()
		comp.emitByte(OpAssert)
	default:
		comp.expression()
		comp.emitByte(OpPop)
	}
}

func (comp *compiler) expression() {
	comp.term()
}

func (comp *compiler) term() {
	comp.factor()

	switch comp.current()._type {
	case TokenPlus:
		comp.advance()
		comp.term()
		comp.emitByte(OpAdd)
	case TokenMinus:
		comp.advance()
		comp.term()
		comp.emitByte(OpSubtract)
	case TokenEof, TokenSemicolon:
		// todo: infinite loop somewhere in here when unrecognized token type
		// maybe instead just require semicolon, maybe this should be an error?
		// Should this collapse into the branch below?
		return
	default:
		return
	}

}

func (comp *compiler) factor() {

	if comp.current()._type == TokenMinus {
		comp.advance()
		comp.factor()
		comp.emitByte(OpNegate)
	} else {
		comp.primary()
	}

	switch comp.current()._type {
	case TokenStar:
		comp.advance()
		comp.factor()
		comp.emitByte(OpMult)
	case TokenSlash:
		comp.advance()
		comp.factor()
		comp.emitByte(OpDivide)
	default:
		return
	}

}

func (comp *compiler) primary() {
	switch comp.current()._type {
	case TokenTrue:
		b := comp.makeConstant(&boolean{true})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	case TokenNumber:
		flt, err := strconv.ParseFloat(
			comp.current().text,
			64,
		)

		if err != nil {
			comp.error(fmt.Sprint("Cannot parse number: ", comp.current().text))
		}

		b := comp.makeConstant(&number{flt})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	default:
		comp.advance()
		comp.error(fmt.Sprint("Unexpected token:", comp.current()))
	}
}

func (comp *compiler) advance() {
	comp.curr += 1
}

func (comp *compiler) consume(tt TokenType) {
	if comp.current()._type != tt {
		comp.error(fmt.Sprint("Expected type ", tt, ", found ", comp.current()._type))
	}

	comp.advance()
}

func (comp *compiler) makeConstant(value Value) byte {
	index := len(comp.chunk.constants)

	// todo: error if too many constants
	comp.chunk.constants = append(comp.chunk.constants, value)

	return byte(index)
}

func (comp *compiler) emitBytes(b1, b2 byte) {
	comp.chunk.bytecode = append(comp.chunk.bytecode, op(b1))
	comp.chunk.bytecode = append(comp.chunk.bytecode, op(b2))
}

func (comp *compiler) emitByte(b byte) {
	comp.chunk.bytecode = append(comp.chunk.bytecode, op(b))
}

func (comp *compiler) emitReturn() {
	last := len(comp.chunk.bytecode) - 1
	if comp.mode == ReplMode && comp.chunk.bytecode[last] == OpPop {
		comp.chunk.bytecode[last] = OpReturn
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
