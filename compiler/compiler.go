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
	OpCall
	OpConstant
	OpCreateTable
	OpDivide
	OpEquals
	OpGetGlobal
	OpGetLocal
	OpGetTable
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
	OpSetLocal
	OpSetTable
	OpInitTable
	OpInsertTable
	OpSubtract
	OpZero

	PrintBytecode bool = true
)

const (
	RunFileMode = iota
	ReplMode
)

type ReturnMode int

type Local struct {
	name  Identifier
	scope int
}

type compiler struct {
	text   []scanner.Token
	curr   int
	chunk  value.Chunk
	name   string
	locals []Local
	scope  int
	err    glerror.GluaErrorChain
	mode   ReturnMode
	parent *compiler
}

type Function struct {
	Chunk value.Chunk
	Name  string
}

func Compile(text []scanner.Token, mode ReturnMode) (Function, glerror.GluaErrorChain) {
	compiler := compiler{
		text:   text,
		curr:   0,
		chunk:  value.Chunk{},
		name:   "",
		locals: []Local{{"", 0}}, // Top-level function has no name
		scope:  0,
		err:    glerror.GluaErrorChain{},
		mode:   mode,
	}

	compiler.compile()

	return compiler.end()
}

func (compiler *compiler) compile() {
	for compiler.current().Type != scanner.TokenEof {
		decl := compiler.declaration()
		if DebugAst {
			var node Node = decl
			PrintTree(&node)
		}
		decl.Emit(compiler)
	}
}

func (compiler *compiler) peek() scanner.Token {
	if compiler.curr+1 >= len(compiler.text) {
		// todo: when pulling one token at a time use the line of the last token
		return scanner.Token{
			Type: scanner.TokenEof,
			Text: "",
			Line: -1,
		}
	}

	return compiler.text[compiler.curr+1]
}

func (compiler *compiler) current() scanner.Token {
	if compiler.curr >= len(compiler.text) {
		return scanner.Token{
			Type: scanner.TokenEof,
			Text: "",
			Line: -1,
		}
	}

	return compiler.text[compiler.curr]
}

func (compiler *compiler) check(tt scanner.TokenType) bool {
	return compiler.current().Type == tt
}

func (compiler *compiler) declaration() Node {
	var state Node
	switch compiler.current().Type {
	case scanner.TokenGlobal:
		state = compiler.global()
	case scanner.TokenLocal:
		state = compiler.local()
	default:
		state = compiler.statement()
	}

	// Lua allows semicolons but they are not required
	if compiler.check(scanner.TokenSemicolon) {
		compiler.consume(scanner.TokenSemicolon)
	}

	return state
}

func (compiler *compiler) global() Node {
	compiler.consume(scanner.TokenGlobal)
	decl := GlobalDeclaration{Identifier(compiler.identifier()), nil}

	if compiler.check(scanner.TokenEqual) {
		compiler.consume(scanner.TokenEqual)
		var assignment = compiler.assignment()
		decl.assignment = &assignment
	}

	return decl
}

func (compiler *compiler) local() Node {
	compiler.consume(scanner.TokenLocal)
	decl := LocalDeclaration{Identifier(compiler.identifier()), nil}

	if compiler.check(scanner.TokenEqual) {
		compiler.consume(scanner.TokenEqual)
		var assignment = compiler.assignment()
		decl.assignment = &assignment
	}

	return decl
}

func (compiler *compiler) statement() Node {
	switch compiler.current().Type {
	case scanner.TokenAssert:
		compiler.consume(scanner.TokenAssert)
		return AssertStatement{
			value: compiler.expression(),
		}
	case scanner.TokenFunction:
		return compiler.function()
	case scanner.TokenWhile:
		return compiler.whileStatement()
	case scanner.TokenDo:
		return compiler.block()
	case scanner.TokenReturn:
		return compiler.returnStatement()
	default:
		return Expression{compiler.expression()}
	}
}

func (compiler *compiler) function() Node {
	compiler.consume(scanner.TokenFunction)

	name := compiler.identifier()

	parameters := compiler.parameters()
	var declarations []Node

	for !compiler.check(scanner.TokenEnd) {
		declarations = append(declarations, compiler.declaration())
	}

	compiler.consume(scanner.TokenEnd)

	var body Node
	if len(declarations) == 1 {
		body = declarations[0]
	} else {
		body = BlockStatement{declarations}
	}

	return FunctionNode{name, parameters, body}
}

func (compiler *compiler) parameters() []Identifier {
	compiler.consume(scanner.TokenLeftParen)

	var identifiers []Identifier
	for !compiler.check(scanner.TokenRightParen) {
		identifiers = append(identifiers, compiler.identifier())
	}

	compiler.consume(scanner.TokenRightParen)

	return identifiers
}

func (compiler *compiler) block() BlockStatement {
	var block BlockStatement
	compiler.consume(scanner.TokenDo)

	for compiler.current().Type != scanner.TokenEnd {
		block.statements = append(block.statements, compiler.declaration())
	}

	compiler.consume(scanner.TokenEnd)
	return block
}

func (compiler *compiler) startScope() {
	compiler.scope += 1
}

func (compiler *compiler) endScope() {
	compiler.scope -= 1

	// Find the first local in a scope above current scope
	var i int
	for i = 0; i < len(compiler.locals); i++ {
		if compiler.locals[i].scope > compiler.scope {
			break
		}
	}

	// Drop the whole list of locals after that
	if i == 0 {
		compiler.locals = nil
	} else {
		compiler.locals = compiler.locals[0:i]
	}
}

func (compiler *compiler) whileStatement() Node {
	compiler.consume(scanner.TokenWhile)

	return WhileStatement{
		condition: compiler.expression(),
		body:      compiler.block(),
	}
}

func (compiler *compiler) returnStatement() Node {
	compiler.consume(scanner.TokenReturn)

	return ReturnStatement{
		value: compiler.expression(),
	}
}

func (compiler *compiler) expression() Node {
	return compiler.assignment()
}

func (compiler *compiler) assignment() Node {
	logicOr := compiler.logicOr()
	if compiler.check(scanner.TokenEqual) {
		compiler.consume(scanner.TokenEqual)
		return logicOr.assign(compiler)
	} else {
		return logicOr
	}
}

func (compiler *compiler) identifier() Identifier {
	ident := compiler.current().Text
	compiler.consume(scanner.TokenIdentifier)
	return Identifier(ident)
}

func (compiler *compiler) logicOr() Node {
	node := compiler.logicAnd()

	if compiler.current().Type != scanner.TokenOr {
		return node
	}

	lor := LogicOr{node, nil}

	for {
		if compiler.check(scanner.TokenOr) {
			compiler.advance()
			lor.or = append(lor.or, compiler.logicAnd())
		} else {
			return node
		}
	}
}

func (compiler *compiler) logicAnd() Node {
	node := compiler.comparison()

	if compiler.current().Type != scanner.TokenAnd {
		return node
	}

	land := LogicAnd{node, nil}

	for {
		if compiler.check(scanner.TokenAnd) {
			compiler.advance()
			land.and = append(land.and, compiler.comparison())
		} else {
			return land
		}
	}
}

func (compiler *compiler) comparison() Node {
	term := compiler.term()

	if !compiler.isComparison() {
		return term
	}

	compare := Comparison{term, nil}

	for compiler.isComparison() {
		token := compiler.current().Type
		compiler.advance()
		compItem := ComparisonItem{
			term:      compiler.term(),
			compareOp: token,
		}
		compare.items = append(compare.items, compItem)
	}

	return compare
}

func (compiler *compiler) isComparison() bool {
	switch compiler.current().Type {
	case scanner.TokenLess, scanner.TokenGreater, scanner.TokenLessEqual,
		scanner.TokenGreaterEqual, scanner.TokenTildeEqual, scanner.TokenEqualEqual:
		return true
	default:
		return false
	}
}

func (compiler *compiler) term() Node {
	factor := compiler.factor()

	if !compiler.isTerm() {
		return factor
	}

	term := Term{factor, nil}

	for compiler.isTerm() {
		token := compiler.current().Type
		compiler.advance()
		termItem := TermItem{
			factor: compiler.factor(),
			termOp: token,
		}
		term.items = append(term.items, termItem)
	}

	return term
}

func (compiler *compiler) isTerm() bool {
	token := compiler.current().Type
	switch token {
	case scanner.TokenMinus, scanner.TokenPlus:
		return true
	default:
		return false
	}
}

func (compiler *compiler) factor() Node {
	unary := compiler.unary()

	if !compiler.isFactor() {
		return unary
	}

	factor := Factor{unary, nil}

	for compiler.isFactor() {
		token := compiler.current().Type
		compiler.advance()
		factorItem := FactorItem{
			unary:    compiler.unary(),
			factorOp: token,
		}
		factor.items = append(factor.items, factorItem)
	}

	return factor
}

func (compiler *compiler) isFactor() bool {
	token := compiler.current().Type
	switch token {
	case scanner.TokenStar, scanner.TokenSlash:
		return true
	default:
		return false
	}
}

func (compiler *compiler) unary() Node {
	switch compiler.current().Type {
	case scanner.TokenMinus:
		compiler.advance()
		return NegateUnary{compiler.unary()}
	case scanner.TokenBang:
		compiler.advance()
		return NotUnary{compiler.unary()}
	default:
		return compiler.exponent()
	}
}

func (compiler *compiler) exponent() Node {
	call := compiler.call()
	if compiler.check(scanner.TokenCaret) {
		compiler.consume(scanner.TokenCaret)
		exp := compiler.call()
		return Exponent{call, &exp}
	} else {
		return call
	}
}

func (compiler *compiler) call() Node {
	primary := compiler.primary()

	for compiler.isCall() {
		switch compiler.current().Type {
		case scanner.TokenDot:
			compiler.consume(scanner.TokenDot)

			attribute := compiler.identifier()

			primary = TableAccessor{
				primary,
				StringPrimary(string(attribute)),
			}
		case scanner.TokenLeftBracket:
			compiler.consume(scanner.TokenLeftBracket)
			attribute := compiler.expression()
			compiler.consume(scanner.TokenRightBracket)

			primary = TableAccessor{
				primary,
				attribute,
			}
		case scanner.TokenLeftParen:
			args := compiler.arguments()

			primary = Call{
				base:      primary,
				arguments: args,
			}
		}
	}

	return primary
}

func (compiler *compiler) isCall() bool {
	switch compiler.current().Type {
	case scanner.TokenDot, scanner.TokenLeftBracket, scanner.TokenLeftParen:
		return true
	default:
		return false
	}
}

func (compiler *compiler) arguments() []Node {
	var args []Node

	compiler.consume(scanner.TokenLeftParen)

	for !compiler.check(scanner.TokenRightParen) {
		args = append(args, compiler.expression())

		if !compiler.check(scanner.TokenRightParen) {
			compiler.consume(scanner.TokenComma)
		}
	}

	compiler.consume(scanner.TokenRightParen)

	return args
}

func (compiler *compiler) primary() Node {
	switch compiler.current().Type {
	case scanner.TokenTrue:
		compiler.advance()
		return BooleanPrimary(true)
	case scanner.TokenFalse:
		compiler.advance()
		return BooleanPrimary(false)
	case scanner.TokenNumber:
		flt, err := strconv.ParseFloat(
			compiler.current().Text,
			64,
		)

		if err != nil {
			compiler.error(fmt.Sprint("Cannot parse number: ", compiler.current().Text))
		}

		compiler.advance()
		return NumberPrimary(flt)
	case scanner.TokenString:
		str := StringPrimary(compiler.current().Text)
		compiler.advance()
		return str
	case scanner.TokenNil:
		n := NilPrimary()
		compiler.advance()
		return n
	case scanner.TokenIdentifier:
		return compiler.variable()
	case scanner.TokenLeftBrace:
		return compiler.tableLiteral()
	case scanner.TokenLeftParen:
		return compiler.grouping()
	default:
		compiler.error(fmt.Sprint("Unexpected token: ", compiler.current()))
		compiler.advance()
		return NilPrimary()
	}
}

func (compiler *compiler) grouping() Node {
	compiler.consume(scanner.TokenLeftParen)
	node := compiler.expression()
	compiler.consume(scanner.TokenRightParen)

	return node
}

func (compiler *compiler) variable() Node {
	name := compiler.current().Text
	compiler.advance()
	return VariablePrimary{Identifier(name)}
}

func (compiler *compiler) getLocal(name Identifier) int {
	for i, local := range compiler.locals {
		if local.name == name {
			return i
		}
	}

	return -1
}

func (compiler *compiler) tableLiteral() TableLiteral {
	compiler.advance()
	var pairs []Node

	// todo: handle unterminated brace
	for compiler.current().Type != scanner.TokenRightBrace {
		pairs = append(pairs, compiler.pair())
	}

	compiler.consume(scanner.TokenRightBrace)

	return TableLiteral{pairs}
}

func (compiler *compiler) pair() Node {
	fmt.Println(compiler.current().Type, compiler.peek().Type)
	switch {
	case compiler.check(scanner.TokenLeftBracket):
		return compiler.literalPair()
	case compiler.check(scanner.TokenIdentifier) && compiler.peek().Type == scanner.TokenEqual:
		return compiler.stringPair()
	default:
		return compiler.value()
	}
}

func (compiler *compiler) literalPair() Node {
	compiler.consume(scanner.TokenLeftBracket)

	expr := compiler.expression()

	compiler.consume(scanner.TokenRightBracket)
	compiler.consume(scanner.TokenEqual)

	value := compiler.expression()

	if compiler.check(scanner.TokenComma) {
		compiler.consume(scanner.TokenComma)
	}

	return LiteralPair{
		key:   expr,
		value: value,
	}
}

func (compiler *compiler) stringPair() Node {
	ident := compiler.identifier()
	compiler.consume(scanner.TokenEqual)

	expr := compiler.expression()

	if compiler.check(scanner.TokenComma) {
		compiler.consume(scanner.TokenComma)
	}

	return StringPair{
		key:   StringPrimary(string(ident)),
		value: expr,
	}
}

func (compiler *compiler) value() Node {
	expr := compiler.expression()

	if compiler.check(scanner.TokenComma) {
		compiler.consume(scanner.TokenComma)
	}

	return Value{expr}
}

func (compiler *compiler) advance() {
	compiler.curr += 1
}

func (compiler *compiler) consume(tt scanner.TokenType) {
	if compiler.current().Type != tt {
		compiler.error(fmt.Sprint("Expected type ", tt, ", found ", compiler.current().Type))
	}

	compiler.advance()
}

// todo: use longer values for more constants
func (compiler *compiler) makeConstant(value value.Value) byte {
	for i, c := range compiler.chunk.Constants {
		if c == value {
			return byte(i)
		}
	}

	index := len(compiler.chunk.Constants)

	// todo: error if too many constants
	compiler.chunk.Constants = append(compiler.chunk.Constants, value)

	return byte(index)
}

func (compiler *compiler) emitBytes(b1, b2 byte) {
	compiler.emitByte(b1)
	compiler.emitByte(b2)
}

func (compiler *compiler) emitByte(b byte) {
	compiler.chunk.Bytecode = append(compiler.chunk.Bytecode, b)
	compiler.chunk.Lines = append(compiler.chunk.Lines, compiler.current().Line)
}

func (compiler *compiler) emitJump(op byte) {
	compiler.emitByte(byte(op))
	compiler.emitBytes(0, 0)
}

func (compiler *compiler) chunkSize() int {
	return len(compiler.chunk.Bytecode)
}

func MergeBytes(upper, lower byte) int {
	return int((int(upper) << 8) | int(lower))
}

func SplitBytes(num int) (upper, lower byte) {
	upper = byte((num >> 8) & 0xff)
	lower = byte(num & 0xff)
	return
}

func (compiler *compiler) patchJump(source, dest int) {
	var dist int
	if dest > source {
		dist = dest - source - 3
	} else {
		dist = source - dest + 3
	}

	upper, lower := SplitBytes(dist)
	compiler.chunk.Bytecode[source+1] = upper
	compiler.chunk.Bytecode[source+2] = lower
}

func (compiler *compiler) emitReturn() {
	last := len(compiler.chunk.Bytecode) - 1
	if compiler.mode == ReplMode && compiler.chunk.Bytecode[last] == OpPop {
		compiler.chunk.Bytecode[last] = OpReturn
	} else {
		compiler.emitBytes(OpNil, OpReturn)
	}
}

func (compiler *compiler) end() (Function, glerror.GluaErrorChain) {
	compiler.emitReturn()

	if !compiler.err.IsEmpty() {
		return Function{}, compiler.err
	}

	// todo: get name from compiler
	function := Function{
		Chunk: compiler.chunk,
		Name:  compiler.name,
	}

	if PrintBytecode {
		DebugPrint(function)
	}

	return function, compiler.err
}

func (compiler *compiler) error(message string) {
	compiler.err.Append(CompileError{
		message: message,
		line:    compiler.current().Line,
	})
}

type CompileError struct {
	message string
	line    int
}

// todo: track line numbers in tokens and print error line
func (ce CompileError) Error() string {
	return fmt.Sprintf("Compile error [line=%d] ---> %s", ce.line, ce.message)
}
