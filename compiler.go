package main

import (
	"fmt"
	"strconv"
)

const (
	OpAdd = iota
	OpConstant
	OpNil
	OpMult
	OpPop
	OpReturn
	OpSubtract
)

type compiler struct {
	text  []Token
	curr  int
	chunk chunk
}

type chunk struct {
	bytecode  []op
	constants []value
}

type Function struct {
	chunk chunk
	name  string
}

func Compile(text []Token) Function {
	compiler := compiler{
		text,
		0,
		chunk{},
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
	comp.expression()
	comp.emitByte(OpPop)
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
	default:
		return
	}

}

func (comp *compiler) factor() {
	comp.primary()

	switch comp.current()._type {
	case TokenStar:
		comp.advance()
		comp.factor()
		comp.emitByte(OpMult)
	default:
		return
	}

}

func (comp *compiler) primary() {
	switch comp.current()._type {
	case TokenNumber:
		flt, err := strconv.ParseFloat(
			comp.current().text,
			64,
		)

		if err != nil {
			panic(fmt.Sprint("Cannot parse number: ", comp.current().text))
		}

		b := comp.makeConstant(&number{flt})
		comp.emitBytes(OpConstant, b)
		comp.advance()
	}
}

func (comp *compiler) advance() {
	comp.curr += 1
}

func (comp *compiler) consume(tt TokenType) {
	if comp.current()._type != tt {
		panic(fmt.Sprint("Expected type ", tt, ", found ", comp.current()._type))
	}

	comp.advance()
}

func (comp *compiler) makeConstant(value value) byte {
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
	comp.emitBytes(OpNil, OpReturn)
}

func (comp *compiler) end() Function {
	comp.emitReturn()

	return Function{
		comp.chunk,
		"",
	}
}
