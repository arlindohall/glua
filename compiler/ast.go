package compiler

import (
	"arlindohall/glua/glerror"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

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

	child := &compiler{
		text:   parent.text,
		curr:   parent.curr,
		chunk:  value.Chunk{},
		name:   string(function.name),
		locals: nil,
		scope:  0,
		err:    glerror.GluaErrorChain{},
		mode:   parent.mode,
		parent: parent,
	}

	child.addLocal(function.name)
	for _, param := range function.parameters {
		child.addLocal(param)
	}

	function.body.Emit(child)

	compiledFunction, err := child.end()

	if !err.IsEmpty() {
		parent.err.AppendAll(&err)
		return
	}

	if parent.scope > 0 {
		parent.locals = append(parent.locals, Local{function.name, parent.scope})
		makeClosure(parent, compiledFunction)
	} else {
		parent.emitByte(OpAssignStart)
		fn := parent.makeConstant(value.StringVal(function.name))
		makeClosure(parent, compiledFunction)
		parent.emitBytes(OpSetGlobal, fn)
		parent.emitByte(OpAssignCleanup)
	}
}

// todo: is there a way to emit the closure without the constant?
// we could key them to the instruction that creates them and look
// the compiled functions in a map??
func makeClosure(compiler *compiler, function Function) {
	c := compiler.makeConstant(value.NewClosure(function.Chunk, function.Name))
	compiler.emitBytes(OpConstant, c)
	compiler.emitByte(OpClosure)

	for _, upvalue := range function.Upvalues {
		compiler.emitBytes(OpCreateUpvalue, byte(upvalue.index))
		compiler.emitByte(toByte(upvalue.isLocal))
	}
}

func toByte(b bool) byte {
	if b {
		return 1
	} else {
		return 0
	}
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
	names  []Identifier
	values []Node
}

func (declaration GlobalDeclaration) Emit(compiler *compiler) {
	compiler.emitByte(OpAssignStart)

	for _, value := range declaration.values {
		value.Emit(compiler)
	}

	for _, name := range declaration.names {
		b := compiler.makeConstant(value.StringVal(name))
		compiler.emitBytes(OpSetGlobal, b)
	}

	compiler.emitByte(OpAssignCleanup)
}

func (declaration GlobalDeclaration) printTree(indent int) {
	printIndent(indent, "GlobalDeclaration")

	for _, name := range declaration.names {
		printIndent(indent+1, string(name))
	}

	for _, value := range declaration.values {
		value.printTree(indent + 1)
	}
}

func (declaration GlobalDeclaration) assign(compiler *compiler) Node {
	// todo should we hook into this for global variables?
	compiler.error("Cannot assign to global variable declaration")
	return declaration
}

type LocalDeclaration struct {
	names  []Identifier
	values []Node
}

func (declaration LocalDeclaration) Emit(compiler *compiler) {
	compiler.emitBytes(OpLocalAllocate, byte(len(declaration.names)))

	for _, value := range declaration.values {
		value.Emit(compiler)
	}

	for _, name := range declaration.names {
		compiler.addLocal(name)
	}

	compiler.emitByte(OpLocalCleanup)
}

func (compiler *compiler) addLocal(name Identifier) {
	local := Local{name, compiler.scope}
	compiler.locals = append(compiler.locals, local)
}

func (declaration LocalDeclaration) printTree(indent int) {
	printIndent(indent, "LocalDeclaration")

	for _, name := range declaration.names {
		printIndent(indent+1, string(name))
	}

	for _, value := range declaration.values {
		value.printTree(indent + 1)
	}
}

func (declaration LocalDeclaration) assign(compiler *compiler) Node {
	// todo should we hook into this for local variables?
	compiler.error("Cannot assign to global variable declaration")
	return declaration
}

type WhileStatement struct {
	condition Node
	body      Node
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

type NumericForStatement struct {
	variable Identifier
	values   []Node
	body     Node
}

// todo: is it better not to inline these? this way there's no
// bounds checks but if we don't inline them, it's less code which
// would also mean less loading from memory
func (statement NumericForStatement) Emit(compiler *compiler) {
	for _, val := range statement.values {
		compiler.startScope()
		compiler.addLocal(statement.variable)
		val.Emit(compiler)

		statement.body.Emit(compiler)

		compiler.emitByte(OpPop)
		compiler.endScope()
	}
}

func (statement NumericForStatement) printTree(indent int) {
	printIndent(indent, "NumericFor")
	printIndent(indent+1, string(statement.variable))

	for _, val := range statement.values {
		val.printTree(indent + 2)
	}

	statement.body.printTree(indent + 1)
}

func (statement NumericForStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to for statement")
	return statement
}

type GenericForStatement struct{}

type IfStatement struct {
	condition      Node
	body           Node
	counterfactual Node
}

func (statement IfStatement) Emit(compiler *compiler) {
	statement.condition.Emit(compiler)

	jumpFromIfFalse := compiler.chunkSize()
	compiler.emitJump(OpJumpIfFalse)

	statement.body.Emit(compiler)
	jumpToIfFalse := compiler.chunkSize()

	if statement.counterfactual != nil {
		compiler.emitByte(OpNil)
		jumpFromIfTrue := compiler.chunkSize()
		compiler.emitJump(OpJumpIfFalse)

		// We jump past the else jump if there's a counterfactual
		jumpToIfFalse = compiler.chunkSize()

		statement.counterfactual.Emit(compiler)
		jumpToIfTrue := compiler.chunkSize()
		compiler.patchJump(jumpFromIfTrue, jumpToIfTrue)
	}

	compiler.patchJump(jumpFromIfFalse, jumpToIfFalse)
}

func (statement IfStatement) printTree(indent int) {
	printIndent(indent, "If")
	statement.condition.printTree(indent + 1)
	statement.body.printTree(indent + 1)

	if statement.counterfactual != nil {
		printIndent(indent+1, "Else")
		statement.counterfactual.printTree(indent + 2)
	}
}

func (statement IfStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to if statement")
	return statement
}

type ReturnStatement struct {
	arity  byte
	values []Node
}

func (statement ReturnStatement) Emit(compiler *compiler) {
	for _, value := range statement.values {
		value.Emit(compiler)
	}

	compiler.emitBytes(OpReturn, statement.arity)
}

func (statement ReturnStatement) printTree(indent int) {
	printIndent(indent, "Return")

	for _, value := range statement.values {
		value.printTree(indent + 1)
	}
}

func (statement ReturnStatement) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to return statement")
	return statement
}

type BlockStatement struct {
	statements []Node
}

func (statement BlockStatement) Emit(compiler *compiler) {
	compiler.startScope()
	for _, st := range statement.statements {
		st.Emit(compiler)
	}
	compiler.endScope()
}

func (compiler *compiler) startScope() {
	compiler.scope += 1
}

func (compiler *compiler) endScope() {
	compiler.scope -= 1

	// Find the first local in a scope above current scope
	var stackTop int
	for stackTop = 0; stackTop < len(compiler.locals); stackTop++ {
		if compiler.locals[stackTop].scope > compiler.scope {
			break
		}
	}

	// function f() x = 1 do y = 2 end return x end
	// stack=[..., Function<f>, 1, 2] <- stack with locals, second is in scope
	//
	// to close the upvalue 2, we want to emit close upvalue up to but not
	// including stackTop
	compiler.emitBytes(OpCloseUpvalues, byte(stackTop))

	// Drop the whole list of locals after that
	if stackTop == 0 {
		compiler.locals = nil
	} else {
		compiler.locals = compiler.locals[0:stackTop]
	}
}

func (statement BlockStatement) printTree(indent int) {
	printIndent(indent, "Block")

	for _, st := range statement.statements {
		st.printTree(indent + 1)
	}
}

func (statement BlockStatement) assign(compiler *compiler) Node {
	// todo: should blocks return their last value and thus assign if it's a table?
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

type MultipleAssignment struct {
	variables []Node
	values    []Node
}

func (assignment MultipleAssignment) Emit(compiler *compiler) {
	compiler.emitByte(OpAssignStart)

	for _, value := range assignment.values {
		value.Emit(compiler)
	}

	for _, variable := range assignment.variables {
		variable.assign(compiler).Emit(compiler)
	}

	compiler.emitByte(OpAssignCleanup)
}

func (assignment MultipleAssignment) printTree(indent int) {
	printIndent(indent, "Assignment")
	printIndent(indent+1, "Variables")

	for _, variable := range assignment.variables {
		variable.printTree(indent + 2)
	}

	printIndent(indent+1, "Values")

	for _, value := range assignment.values {
		value.printTree(indent + 2)
	}
}

func (assignment MultipleAssignment) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to assignment")
	return assignment
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
	name Identifier
}

func (assignment VariableAssignment) Emit(compiler *compiler) {
	// todo: determine local/upvalue/global when building AST
	local := compiler.getLocal(assignment.name)
	if local != -1 {
		compiler.emitBytes(OpSetLocal, byte(local))
		return
	}

	upvalue := compiler.getUpvalue(assignment.name)
	if upvalue != -1 {
		compiler.emitBytes(OpSetUpvalue, byte(upvalue))
		return
	}

	name := compiler.makeConstant(value.StringVal(assignment.name))
	compiler.emitBytes(OpSetGlobal, name)
}

func (assignment VariableAssignment) printTree(indent int) {
	printIndent(indent, assignment.name)
}

func (assignment VariableAssignment) assign(compiler *compiler) Node {
	panic("Cannot assign to assignment")
}

type TableAssignment struct {
	table     Node
	attribute Node
}

func (assignment TableAssignment) Emit(compiler *compiler) {
	assignment.table.Emit(compiler)
	assignment.attribute.Emit(compiler)
	compiler.emitByte(OpSetTable)
}

func (assignment TableAssignment) printTree(indent int) {
	printIndent(indent, "AssignTable")
	assignment.table.printTree(indent + 1)
	assignment.attribute.printTree(indent + 1)
}

func (assignment TableAssignment) assign(compiler *compiler) Node {
	compiler.error("Cannot assign to assignment")
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
	return TableAssignment(accessor)
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
		case scanner.TokenTildeEqual:
			compiler.emitBytes(OpEquals, OpNot)
		case scanner.TokenLess:
			compiler.emitByte(OpLess)
		case scanner.TokenLessEqual:
			compiler.emitBytes(OpGreater, OpNot)
		case scanner.TokenGreater:
			compiler.emitByte(OpGreater)
		case scanner.TokenGreaterEqual:
			compiler.emitBytes(OpLess, OpNot)
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
	base         Node
	arguments    []Node
	isAssignment bool
}

func (call *Call) Emit(compiler *compiler) {
	call.base.Emit(compiler)

	arity := 0
	for _, arg := range call.arguments {
		arg.Emit(compiler)
		arity++
	}

	// todo: audit places where ints are downcast for overflow
	// -> arity, locals, upvalues
	compiler.emitBytes(OpCall, byte(arity))
	compiler.emitByte(toByte(call.isAssignment))
}

func (call *Call) printTree(indent int) {
	printIndent(indent, "Call")
	call.base.printTree(indent + 1)

	if len(call.arguments) > 0 {
		printIndent(indent+1, "Arguments")
	}

	for _, arg := range call.arguments {
		arg.printTree(indent + 2)
	}
}

// todo: is this too hacky?
func (call *Call) assign(compiler *compiler) Node {
	if compiler == nil {
		call.isAssignment = true
	} else {
		compiler.error("Cannot assign to function call")
	}
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
	if local != -1 {
		compiler.emitBytes(OpGetLocal, byte(local))
		return
	}

	upvalue := compiler.getUpvalue(name)
	if upvalue != -1 {
		compiler.emitBytes(OpGetUpvalue, byte(upvalue))
		return
	}

	constant := compiler.makeConstant(value.StringVal(string(name)))
	compiler.emitBytes(OpGetGlobal, constant)

}

// todo: encode block scope and locals into types so they can be used
// when printing not jus when emitting
func (primary VariablePrimary) printTree(indent int) {
	printIndent(indent, fmt.Sprintf("Identifier/%s", string(primary.name)))
}

func (primary VariablePrimary) assign(compiler *compiler) Node {
	return VariableAssignment(primary)
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
