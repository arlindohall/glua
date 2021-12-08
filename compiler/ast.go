package compiler

import (
	"arlindohall/glua/glerror"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

const DebugAst = true

type Node interface {
	Emit(compiler *compiler)
	printTree(indent int)
	assign(compiler *compiler) Node
}

func PrintTree(node *Node) {
	(*node).printTree(0)
}

type FunctionNode struct {
	name       Identifier
	parameters []Identifier
	body       Node
}

func (function FunctionNode) Emit(parent *compiler) {
	var parameters []Local

	for _, param := range function.parameters {
		parameters = append(parameters, Local{param, 0})
	}

	child := &compiler{
		text:   parent.text,
		curr:   parent.curr,
		chunk:  value.Chunk{},
		name:   string(function.name),
		locals: parameters,
		scope:  0,
		err:    glerror.GluaErrorChain{},
		mode:   parent.mode,
		parent: parent,
	}

	function.body.Emit(child)

	result, err := child.end()

	if !err.IsEmpty() {
		parent.err.AppendAll(&err)
		return
	}

	parent.locals = append(parent.locals, Local{function.name, parent.scope})
	c := parent.makeConstant(value.NewClosure(result.Chunk, string(function.name)))
	parent.emitBytes(OpConstant, c)
}

func (function FunctionNode) printTree(indent int) {
	printIndent(indent, "Function")
	printIndent(indent+1, function.name)

	printIndent(indent+1, "Parameters")
	for _, p := range function.parameters {
		printIndent(indent+2, p)
	}

	printIndent(indent+1, "Body")
	function.body.printTree(indent + 2)
}

func (function FunctionNode) assign(compiler *compiler) Node {
	panic("Cannot assign to function declaration")
}

type GlobalDeclaration struct {
	name       Identifier
	assignment *Node
}

func (declaration GlobalDeclaration) Emit(compiler *compiler) {
	b := compiler.makeConstant(value.StringVal(declaration.name))
	if declaration.assignment != nil {
		(*declaration.assignment).Emit(compiler)
		compiler.emitBytes(OpSetGlobal, b)
		compiler.emitByte(OpPop)
	} else {
		compiler.emitByte(OpNil)
		compiler.emitBytes(OpSetGlobal, b)
		compiler.emitByte(OpPop)
	}
}

func (declaration GlobalDeclaration) printTree(indent int) {
	printIndent(indent, "GlobalDeclaration")
	printIndent(indent+1, declaration.name)

	if declaration.assignment != nil {
		(*declaration.assignment).printTree(indent + 1)
	}
}

func (declaration GlobalDeclaration) assign(compiler *compiler) Node {
	// todo should we hook into this for global variables?
	compiler.error("Cannot assign to global variable declaration")
	return declaration
}

type LocalDeclaration struct {
	name       Identifier
	assignment *Node
}

func (declaration LocalDeclaration) Emit(compiler *compiler) {
	// todo: more than one local
	var local Local
	if declaration.assignment != nil {
		local = Local{declaration.name, compiler.scope}
		(*declaration.assignment).Emit(compiler)
	} else {
		local = Local{declaration.name, compiler.scope}
		compiler.emitByte(OpNil)
	}
	compiler.locals = append(compiler.locals, local)
}

func (declaration LocalDeclaration) printTree(indent int) {
	printIndent(indent, "LocalDeclaration")
	printIndent(indent+1, declaration.name)

	if declaration.assignment != nil {
		(*declaration.assignment).printTree(indent + 1)
	}
}

func (declaration LocalDeclaration) assign(compiler *compiler) Node {
	// todo should we hook into this for local variables?
	compiler.error("Cannot assign to global variable declaration")
	return declaration
}

type WhileStatement struct {
	condition Node
	body      BlockStatement
}

func (statement WhileStatement) Emit(compiler *compiler) {
	loopTo := compiler.chunkSize()
	statement.condition.Emit(compiler)

	jumpFrom := compiler.chunkSize()
	compiler.emitJump(OpJumpIfFalse)

	statement.body.Emit(compiler)

	loopFrom := compiler.chunkSize()
	compiler.emitJump(OpLoop)
	jumpTo := compiler.chunkSize()

	compiler.patchJump(loopFrom, loopTo)
	compiler.patchJump(jumpFrom, jumpTo)
}

func (statement WhileStatement) printTree(indent int) {
	printIndent(indent, "While")
	statement.condition.printTree(indent + 1)

	statement.body.printTree(indent + 1)
}

func (statement WhileStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to while statement")
	return statement
}

type ReturnStatement struct {
	value Node
}

func (statement ReturnStatement) Emit(compiler *compiler) {
	statement.value.Emit(compiler)
	compiler.emitByte(OpReturn)
}

func (statement ReturnStatement) printTree(indent int) {
	printIndent(indent, "Return")
	statement.value.printTree(indent + 1)
}

func (statement ReturnStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to return statement")
	return statement
}

type BlockStatement struct {
	statements []Node
}

// todo: block scope
func (statement BlockStatement) Emit(compiler *compiler) {
	compiler.startScope()
	for _, st := range statement.statements {
		st.Emit(compiler)
	}
	compiler.endScope()
}

func (statement BlockStatement) printTree(indent int) {
	printIndent(indent, "Block")

	for _, st := range statement.statements {
		st.printTree(indent + 1)
	}
}

func (statement BlockStatement) assign(compiler *compiler) Node {
	// todo should blocks return their last value and thus assign if it's a table?
	compiler.error("Cannot assign to block statement")
	return statement
}

type AssertStatement struct {
	value Node
}

func (statement AssertStatement) Emit(compiler *compiler) {
	statement.value.Emit(compiler)
	compiler.emitByte(OpAssert)
}

func (statement AssertStatement) printTree(indent int) {
	printIndent(indent, "Assert")
	statement.value.printTree(indent + 1)
}

func (statement AssertStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to expression")
	return statement
}

type Expression struct {
	expression Node
}

func (statement Expression) Emit(compiler *compiler) {
	statement.expression.Emit(compiler)
	compiler.emitByte(OpPop)
}

func (statement Expression) printTree(indent int) {
	printIndent(indent, "Expression")
	statement.expression.printTree(indent + 1)
}

func (statement Expression) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to expression")
	return statement
}

type VariableAssignment struct {
	name  Identifier
	value Node
}

func (assignment VariableAssignment) Emit(compiler *compiler) {
	assignment.value.Emit(compiler)

	local := compiler.getLocal(assignment.name)
	if local == -1 {
		name := compiler.makeConstant(value.StringVal(assignment.name))
		compiler.emitBytes(OpSetGlobal, name)
	} else {
		compiler.emitBytes(OpSetLocal, byte(local))
	}
}

func (assignment VariableAssignment) printTree(indent int) {
	printIndent(indent, "Assign")
	printIndent(indent+1, assignment.name)
	assignment.value.printTree(indent + 1)
}

func (assignment VariableAssignment) assign(compiler *compiler) Node {
	target := assignment.name
	intermediate := assignment.value

	value := intermediate.assign(compiler)

	return VariableAssignment{
		target,
		value,
	}
}

type TableAssignment struct {
	table     Node
	attribute Node
	value     Node
}

func (assignment TableAssignment) Emit(compiler *compiler) {
	assignment.table.Emit(compiler)
	assignment.attribute.Emit(compiler)
	assignment.value.Emit(compiler)
	compiler.emitByte(OpSetTable)
}

func (assignment TableAssignment) printTree(indent int) {
	printIndent(indent, "AssignTable")
	assignment.table.printTree(indent + 1)
	assignment.attribute.printTree(indent + 1)
	assignment.value.printTree(indent + 1)
}

func (assignment TableAssignment) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to assignment (yet)")
	return assignment
}

type TableAccessor struct {
	table     Node
	attribute Node
}

func (accessor TableAccessor) Emit(compiler *compiler) {
	accessor.table.Emit(compiler)
	accessor.attribute.Emit(compiler)

	compiler.emitByte(OpGetTable)
}

func (accessor TableAccessor) printTree(indent int) {
	printIndent(indent, "TableGet")
	accessor.table.printTree(indent + 1)
	accessor.attribute.printTree(indent + 1)
}

func (accessor TableAccessor) assign(compiler *compiler) Node {
	return TableAssignment{
		table:     accessor.table,
		attribute: accessor.attribute,
		value:     compiler.expression(),
	}
}

type LogicOr struct {
	value Node
	or    []Node
}

// todo: short circuit or with a jump
func (logicOr LogicOr) Emit(compiler *compiler) {
	logicOr.value.Emit(compiler)
	for _, la := range logicOr.or {
		la.Emit(compiler)
		compiler.emitByte(OpOr)
	}
}

func (logicOr LogicOr) printTree(indent int) {
	if len(logicOr.or) == 0 {
		logicOr.value.printTree(indent)
		return
	}

	printIndent(indent, "Or")
	logicOr.value.printTree(indent + 1)
	for _, or := range logicOr.or {
		or.printTree(indent + 1)
	}
}

func (logicOr LogicOr) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to logical or")
	return logicOr
}

type LogicAnd struct {
	value Node
	and   []Node
}

func (logicAnd LogicAnd) Emit(compiler *compiler) {
	logicAnd.value.Emit(compiler)
	for _, comp := range logicAnd.and {
		comp.Emit(compiler)
		compiler.emitByte(OpAnd)
	}
}

func (logicAnd LogicAnd) printTree(indent int) {
	if len(logicAnd.and) == 0 {
		logicAnd.value.printTree(indent)
		return
	}

	printIndent(indent, "And")
	logicAnd.value.printTree(indent + 1)
	for _, comp := range logicAnd.and {
		comp.printTree(indent + 1)
	}
}

func (logicAnd LogicAnd) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to logical and")
	return logicAnd
}

type Comparison struct {
	term  Node
	items []ComparisonItem
}

// todo: could return second rather than pop but that would not
// be compatible with Lua
func (comparison Comparison) Emit(compiler *compiler) {
	comparison.term.Emit(compiler)
	for _, ci := range comparison.items {
		ci.term.Emit(compiler)
		switch ci.compareOp {
		case scanner.TokenEqualEqual:
			compiler.emitByte(OpEquals)
		case scanner.TokenLess:
			compiler.emitByte(OpLessThan)
		default:
			compiler.error(fmt.Sprint("Unknown comparator operator: ", ci.compareOp))
		}
	}
}

func (comparison Comparison) printTree(indent int) {
	if len(comparison.items) == 0 {
		comparison.term.printTree(indent)
		return
	}

	printIndent(indent, comparison.items[0].compareOp)
	comparison.term.printTree(indent + 1)
	Comparison{
		comparison.items[0].term,
		comparison.items[1:],
	}.printTree(indent + 1)
}

func (comparison Comparison) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to comparison")
	return comparison
}

type ComparisonItem struct {
	compareOp scanner.TokenType
	term      Node
}

type Term struct {
	factor Node
	items  []TermItem
}

func (term Term) Emit(compiler *compiler) {
	term.factor.Emit(compiler)
	for _, ti := range term.items {
		ti.factor.Emit(compiler)
		switch ti.termOp {
		case scanner.TokenPlus:
			compiler.emitByte(OpAdd)
		case scanner.TokenMinus:
			compiler.emitByte(OpSubtract)
		default:
			compiler.error(fmt.Sprint("Unknown term operator: ", ti.termOp))
		}
	}
}

func (term Term) printTree(indent int) {
	if len(term.items) == 0 {
		term.factor.printTree(indent)
		return
	}

	printIndent(indent, term.items[0].termOp)
	term.factor.printTree(indent + 1)

	Term{
		term.items[0].factor,
		term.items[1:],
	}.printTree(indent + 1)
}

func (term Term) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to term")
	return term
}

type TermItem struct {
	termOp scanner.TokenType
	factor Node
}

type Factor struct {
	unary Node
	items []FactorItem
}

func (factor Factor) Emit(compiler *compiler) {
	factor.unary.Emit(compiler)
	for _, u := range factor.items {
		u.unary.Emit(compiler)
		switch u.factorOp {
		case scanner.TokenStar:
			compiler.emitByte(OpMult)
		case scanner.TokenSlash:
			compiler.emitByte(OpDivide)
		default:
			compiler.error(fmt.Sprint("Unkown factor operator: ", u.factorOp))
		}
	}
}

func (factor Factor) printTree(indent int) {
	if len(factor.items) == 0 {
		factor.unary.printTree(indent)
		return
	}

	printIndent(indent, factor.items[0].factorOp)
	factor.unary.printTree(indent + 1)

	Factor{
		factor.items[0].unary,
		factor.items[1:],
	}.printTree(indent + 1)
}

func (factor Factor) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to factor")
	return factor
}

type FactorItem struct {
	factorOp scanner.TokenType
	unary    Node
}

type NegateUnary struct {
	unary Node
}

func (unary NegateUnary) Emit(compiler *compiler) {
	unary.unary.Emit(compiler)
	compiler.emitByte(OpNegate)
}

func (unary NegateUnary) printTree(indent int) {
	printIndent(indent, "Negate")
	unary.unary.printTree(indent + 1)
}

func (unary NegateUnary) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to unary")
	return unary
}

type NotUnary struct {
	unary Node
}

func (unary NotUnary) Emit(compiler *compiler) {
	unary.unary.Emit(compiler)
	compiler.emitByte(OpNot)
}

func (unary NotUnary) printTree(indent int) {
	printIndent(indent, "Not")
	unary.unary.printTree(indent + 1)
}

func (unary NotUnary) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to unary")
	return unary
}

type Exponent struct {
	base Node
	exp  *Node
}

func (exponent Exponent) Emit(compiler *compiler) {
	exponent.base.Emit(compiler)

	if exponent.exp != nil {
		(*exponent.exp).Emit(compiler)
		panic("todo exponentiation")
	}
}

func (exponent Exponent) printTree(indent int) {
	if exponent.exp == nil {
		exponent.base.printTree(indent)
	} else {
		printIndent(indent, "Exp")
		exponent.base.printTree(indent + 1)
		(*exponent.exp).printTree(indent + 1)
	}
}

func (exponent Exponent) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to exponent")
	return exponent
}

type Call struct {
	base      Node
	arguments []Node
}

func (call Call) Emit(compiler *compiler) {
	call.base.Emit(compiler)

	arity := 0
	for _, arg := range call.arguments {
		arg.Emit(compiler)
		arity++
	}

	c := compiler.makeConstant(value.Number(arity))
	compiler.emitBytes(OpConstant, c)

	compiler.emitByte(OpCall)
}

func (call Call) printTree(indent int) {
	printIndent(indent, "Call")
	call.base.printTree(indent + 1)

	if len(call.arguments) > 0 {
		printIndent(indent+1, "Arguments")
	}

	for _, arg := range call.arguments {
		arg.printTree(indent + 2)
	}
}

func (call Call) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to function call")
	return call
}

type LiteralPrimary struct {
	value value.Value
}

func (primary LiteralPrimary) Emit(compiler *compiler) {
	b := compiler.makeConstant(primary.value)
	compiler.emitBytes(OpConstant, b)
}

func (primary LiteralPrimary) printTree(indent int) {
	printIndent(indent, primary.value)
}

func (primary LiteralPrimary) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to literal")
	return primary
}

func NumberPrimary(f float64) LiteralPrimary {
	return LiteralPrimary{
		value: value.Number(f),
	}
}

func BooleanPrimary(b bool) LiteralPrimary {
	return LiteralPrimary{
		value: value.Boolean(b),
	}
}

func StringPrimary(s string) LiteralPrimary {
	return LiteralPrimary{
		value: value.StringVal(s),
	}
}

func NilPrimary() LiteralPrimary {
	return LiteralPrimary{
		value: value.Nil{},
	}
}

type VariablePrimary struct {
	name Identifier
}

func (primary VariablePrimary) Emit(compiler *compiler) {
	name := primary.name

	local := compiler.getLocal(name)
	if local == -1 {
		constant := compiler.makeConstant(value.StringVal(string(name)))
		compiler.emitBytes(OpGetGlobal, constant)
	} else {
		compiler.emitBytes(OpGetLocal, byte(local))
	}
}

func (primary VariablePrimary) printTree(indent int) {
	printIndent(indent, fmt.Sprintf("Global/%s", string(primary.name)))
}

func (primary VariablePrimary) assign(compiler *compiler) Node {
	return VariableAssignment{
		name:  primary.name,
		value: compiler.expression(),
	}
}

type TableLiteral struct {
	entries []Node
}

func (literal TableLiteral) Emit(compiler *compiler) {
	compiler.emitByte(OpCreateTable)

	for _, ent := range mapPairs(literal.entries) {
		ent.Emit(compiler)
	}

	for _, ent := range valuePairs(literal.entries) {
		ent.Emit(compiler)
	}
}

func valuePairs(pairs []Node) []Node {
	var mapPairs []Node

	for _, p := range pairs {
		switch p.(type) {
		case Value:
			mapPairs = append(mapPairs, p)
		}
	}

	return mapPairs
}

func mapPairs(pairs []Node) []Node {
	var mapPairs []Node

	for _, p := range pairs {
		switch p.(type) {
		case Value:
		default:
			mapPairs = append(mapPairs, p)
		}
	}

	return mapPairs
}

func (literal TableLiteral) printTree(indent int) {
	printIndent(indent, "Table")

	for _, entry := range literal.entries {
		entry.printTree(indent + 1)
	}
}

func (literal TableLiteral) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to table literal")
	return literal
}

type Value struct {
	value Node
}

func (val Value) Emit(compiler *compiler) {
	val.value.Emit(compiler)
	compiler.emitByte(OpInsertTable)
}

func (val Value) printTree(indent int) {
	printIndent(indent, "Value")
	val.value.printTree(indent + 1)
}

func (val Value) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to table value")
	return val
}

type StringPair struct {
	key   Node
	value Node
}

func (pair StringPair) Emit(compiler *compiler) {
	pair.key.Emit(compiler)
	pair.value.Emit(compiler)
	compiler.emitByte(OpInitTable)
}

func (pair StringPair) printTree(indent int) {
	printIndent(indent, "StringPair")
	pair.key.printTree(indent + 1)
	pair.value.printTree(indent + 1)
}

func (pair StringPair) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to table literal pair")
	return pair
}

type LiteralPair struct {
	key   Node
	value Node
}

func (pair LiteralPair) Emit(compiler *compiler) {
	pair.key.Emit(compiler)
	pair.value.Emit(compiler)
	compiler.emitByte(OpInitTable)
}

func (pair LiteralPair) printTree(indent int) {
	printIndent(indent, "LiteralPair")
	pair.key.printTree(indent + 1)
	pair.value.printTree(indent + 1)
}

func (pair LiteralPair) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to table literal pair")
	return pair
}

type Identifier string

func printIndent(indent int, node interface{}) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(os.Stderr, "  ")
	}
	fmt.Fprintln(os.Stderr, node)
}
