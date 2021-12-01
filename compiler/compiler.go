package compiler

import (
	"arlindohall/glua/glerror"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
	"strconv"
)

const (
	OpAdd = iota
	OpAssert
	OpAnd
	OpConstant
	OpCreateTable
	OpDivide
	OpEquals
	OpGetGlobal
	OpJumpIfFalse
	OpLessThan
	OpLoop
	OpMult
	OpNegate
	OpNil
	OpNot
	OpOr
	OpPop
	OpReturn
	OpSetGlobal
	OpSetTable
	OpInsertTable
	OpSubtract
	OpZero
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
	err   glerror.GluaErrorChain
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

func Compile(text []scanner.Token, mode ReturnMode) (Function, glerror.GluaErrorChain) {
	compiler := compiler{
		text,
		0,
		Chunk{},
		glerror.GluaErrorChain{},
		mode,
	}

	compiler.compile()

	return compiler.end()
}

func (comp *compiler) compile() {
	for comp.current().Type != scanner.TokenEof {
		decl := comp.declaration()
		if DebugAst {
			decl.PrintTree()
		}
		decl.EmitDeclaration(comp)
	}
}

func (comp *compiler) peek() scanner.Token {
	if comp.curr+1 >= len(comp.text) {
		return scanner.Token{
			Type: scanner.TokenEof,
			Text: "",
		}
	}

	return comp.text[comp.curr+1]
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

func (comp *compiler) declaration() Declaration {
	var state Declaration
	if comp.current().Type == scanner.TokenGlobal {
		state = comp.global()
	} else {
		state = StatementDeclaration{comp.statement()}
	}

	// Lua allows semicolons but they are not required
	if comp.current().Type == scanner.TokenSemicolon {
		comp.consume(scanner.TokenSemicolon)
	}

	return state
}

func (comp *compiler) global() Declaration {
	comp.advance()
	decl := GlobalDeclaration{comp.current().Text, nil}

	comp.advance()
	if comp.current().Type == scanner.TokenEqual {
		comp.advance()
		decl.assignment = comp.expression()
	}

	return decl
}

func (comp *compiler) statement() Statement {
	switch comp.current().Type {
	case scanner.TokenAssert:
		comp.advance()
		return AssertStatement{
			value: comp.expression(),
		}
	case scanner.TokenWhile:
		return comp.whileStatement()
	default:
		return ExpressionStatement{comp.expression()}
	}
}

func (comp *compiler) block() BlockStatement {
	var block BlockStatement

	if comp.current().Type != scanner.TokenDo {
		comp.advance()
		block.statements = append(block.statements, comp.statement())
		return block
	}

	comp.consume(scanner.TokenDo)

	for comp.current().Type != scanner.TokenEnd {
		block.statements = append(block.statements, comp.statement())
	}

	comp.consume(scanner.TokenEnd)

	return block
}

func (comp *compiler) whileStatement() Statement {
	comp.advance()

	return WhileStatement{
		comp.expression(),
		comp.block(),
	}
}

func (comp *compiler) expression() Expression {
	return comp.assignment()
}

func (comp *compiler) assignment() Expression {
	if comp.current().Type == scanner.TokenIdentifier && comp.peek().Type == scanner.TokenEqual {
		name := comp.current().Text
		comp.advance()
		comp.advance()

		return Assignment{
			name,
			comp.expression(),
		}
	} else {
		return comp.logicOr()
	}
}

func (comp *compiler) logicOr() LogicOr {
	lor := LogicOr{comp.logicAnd(), nil}

	for {
		if comp.current().Type == scanner.TokenOr {
			comp.advance()
			lor.or = append(lor.or, comp.logicAnd())
		} else {
			return lor
		}
	}
}

func (comp *compiler) logicAnd() LogicAnd {
	land := LogicAnd{comp.comparison(), nil}

	for {
		if comp.current().Type == scanner.TokenAnd {
			comp.advance()
			land.and = append(land.and, comp.comparison())
		} else {
			return land
		}
	}
}

func (comp *compiler) comparison() Comparison {
	compare := Comparison{comp.term(), nil}

	for {
		token := comp.current().Type
		switch token {
		case scanner.TokenLess, scanner.TokenGreater, scanner.TokenLessEqual, scanner.TokenGreaterEqual, scanner.TokenTildeEqual, scanner.TokenEqualEqual:
			comp.advance()
			compItem := ComparisonItem{
				term:      comp.term(),
				compareOp: token,
			}
			compare.items = append(compare.items, compItem)
		default:
			return compare
		}
	}
}

func (comp *compiler) term() Term {
	term := Term{comp.factor(), nil}

	for {
		token := comp.current().Type
		switch token {
		case scanner.TokenMinus, scanner.TokenPlus:
			comp.advance()
			termItem := TermItem{
				factor: comp.factor(),
				termOp: token,
			}
			term.items = append(term.items, termItem)
		default:
			return term
		}
	}
}

func (comp *compiler) factor() Factor {
	factor := Factor{comp.unary(), nil}

	for {
		token := comp.current().Type
		switch token {
		case scanner.TokenStar, scanner.TokenSlash:
			comp.advance()
			factorItem := FactorItem{
				unary:    comp.unary(),
				factorOp: token,
			}
			factor.items = append(factor.items, factorItem)
		default:
			return factor
		}
	}

}

func (comp *compiler) unary() Unary {
	switch comp.current().Type {
	case scanner.TokenMinus:
		comp.advance()
		return NegateUnary{comp.unary()}
	case scanner.TokenBang:
		comp.advance()
		return NotUnary{comp.unary()}
	default:
		return BaseUnary{comp.exponent()}
	}
}

func (comp *compiler) exponent() Exponent {
	if comp.current().Type == scanner.TokenCaret {
		comp.advance()
		return Exponent{comp.primary(), nil}
	} else {
		return Exponent{comp.primary(), nil}
	}
}

func (comp *compiler) primary() Primary {
	switch comp.current().Type {
	case scanner.TokenTrue:
		comp.advance()
		return BooleanPrimary(true)
	case scanner.TokenFalse:
		comp.advance()
		return BooleanPrimary(false)
	case scanner.TokenNumber:
		flt, err := strconv.ParseFloat(
			comp.current().Text,
			64,
		)

		if err != nil {
			comp.error(fmt.Sprint("Cannot parse number: ", comp.current().Text))
		}

		comp.advance()
		return NumberPrimary(flt)
	case scanner.TokenString:
		str := StringPrimary(comp.current().Text)
		comp.advance()
		return str
	case scanner.TokenNil:
		n := NilPrimary()
		comp.advance()
		return n
	case scanner.TokenIdentifier:
		return comp.variable()
	case scanner.TokenLeftBrace:
		return comp.tableLiteral()
	default:
		comp.error(fmt.Sprint("Unexpected token: ", comp.current()))
		comp.advance()
		return NilPrimary()
	}
}

func (comp *compiler) variable() Primary {
	name := comp.current().Text
	comp.advance()
	return GlobalPrimary(name)
}

func (comp *compiler) tableLiteral() TableLiteral {
	comp.advance()
	var pairs []Pair

	// todo: handle unterminated brace
	for comp.current().Type != scanner.TokenRightBrace {
		pairs = append(pairs, comp.pair())
	}

	comp.consume(scanner.TokenRightBrace)

	return TableLiteral{pairs}
}

func (comp *compiler) pair() Pair {
	switch {
	case comp.current().Type == scanner.TokenLeftBracket:
		return comp.literalPair()
	case comp.current().Type == scanner.TokenIdentifier && comp.peek().Type == scanner.TokenEqual:
		return comp.stringPair()
	default:
		return comp.value()
	}
}

func (comp *compiler) literalPair() Pair {
	comp.consume(scanner.TokenLeftBrace)

	expr := comp.expression()

	comp.consume(scanner.TokenRightBracket)
	comp.consume(scanner.TokenEqual)

	value := comp.expression()

	if comp.current().Type == scanner.TokenComma {
		comp.consume(scanner.TokenComma)
	}

	return LiteralPair{
		key:   expr,
		value: value,
	}
}

func (comp *compiler) stringPair() Pair {
	ident := comp.current().Text

	comp.consume(scanner.TokenEqual)

	expr := comp.expression()

	if comp.peek().Type == scanner.TokenComma {
		comp.consume(scanner.TokenComma)
	}

	return StringPair{
		key:   StringPrimary(ident),
		value: expr,
	}
}

func (comp *compiler) value() Pair {
	expr := comp.expression()

	if comp.current().Type == scanner.TokenComma {
		comp.consume(scanner.TokenComma)
	}

	return Value{expr}
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
	for i, c := range comp.chunk.Constants {
		if c == value {
			return byte(i)
		}
	}

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

func (comp *compiler) emitJump(op Op) {
	comp.emitByte(byte(op))
	comp.emitBytes(0, 0)
}

func (comp *compiler) chunkSize() int {
	return len(comp.chunk.Bytecode)
}

func MergeBytes(upper, lower byte) int {
	return int((int(upper) << 8) | int(lower))
}

func SplitBytes(num int) (upper, lower byte) {
	upper = byte((num >> 8) & 0xff)
	lower = byte(num & 0xff)
	return
}

func (comp *compiler) patchJump(source, dest int) {
	var dist int
	if dest > source {
		dist = dest - source - 3
	} else {
		dist = source - dest + 3
	}

	upper, lower := SplitBytes(dist)
	comp.chunk.Bytecode[source+1] = Op(upper)
	comp.chunk.Bytecode[source+2] = Op(lower)
}

func (comp *compiler) emitReturn() {
	last := len(comp.chunk.Bytecode) - 1
	if comp.mode == ReplMode && comp.chunk.Bytecode[last] == OpPop {
		comp.chunk.Bytecode[last] = OpReturn
	} else {
		comp.emitBytes(OpNil, OpReturn)
	}
}

func (comp *compiler) end() (Function, glerror.GluaErrorChain) {
	comp.emitReturn()

	if !comp.err.IsEmpty() {
		return Function{}, comp.err
	}

	return Function{
		comp.chunk,
		"",
	}, comp.err
}

func (comp *compiler) error(message string) {
	comp.err.Append(CompileError{message})
}

type CompileError struct {
	message string
}

// todo: track line numbers in tokens and print error line
func (ce CompileError) Error() string {
	return fmt.Sprintf("Compile error ---> %s", ce.message)
}
