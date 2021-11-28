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
	OpAnd
	OpConstant
	OpDivide
	OpEquals
	OpMult
	OpNegate
	OpNil
	OpNot
	OpOr
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
		decl := comp.declaration()
		if DebugAst {
			decl.PrintTree()
		}
		decl.EmitDeclaration(comp)
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

func (comp *compiler) declaration() Declaration {
	state := StatementDeclaration{comp.statement()}

	// Lua allows semicolons but they are not required
	if comp.current().Type == scanner.TokenSemicolon {
		comp.consume(scanner.TokenSemicolon)
	}

	return state
}

func (comp *compiler) statement() Statement {
	switch comp.current().Type {
	case scanner.TokenAssert:
		comp.advance()
		return AssertStatement{
			value: comp.expression(),
		}
	default:
		return ExpressionStatement{comp.expression()}
	}
}

func (comp *compiler) expression() Expression {
	return comp.logicOr()
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
	default:
		comp.error(fmt.Sprint("Unexpected token:", comp.current()))
		comp.advance()
		return NilPrimary()
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
