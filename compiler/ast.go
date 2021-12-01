package compiler

import (
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

const DebugAst = true

type Declaration interface {
	EmitDeclaration(*compiler)
	PrintTree()
}

type GlobalDeclaration struct {
	name       string
	assignment Expression
}

func (gd GlobalDeclaration) EmitDeclaration(c *compiler) {
	b := c.makeConstant(value.StringVal(gd.name))
	if gd.assignment != nil {
		gd.assignment.EmitExpression(c)
		c.emitBytes(OpSetGlobal, b)
		c.emitByte(OpPop)
	} else {
		c.emitByte(OpNil)
		c.emitBytes(OpSetGlobal, b)
		c.emitByte(OpPop)
	}
}

func (gd GlobalDeclaration) PrintTree() {
	fmt.Fprintln(os.Stderr, "GlobalDeclaration")
	printIndent(1)
	fmt.Fprintln(os.Stderr, gd.name)

	if gd.assignment != nil {
		gd.assignment.PrintTree(1)
	}
}

type StatementDeclaration struct {
	statement Statement
}

func (sd StatementDeclaration) EmitDeclaration(c *compiler) {
	sd.statement.EmitStatement(c)
}

func (sd StatementDeclaration) PrintTree() {
	fmt.Fprintln(os.Stderr, "Statement")
	sd.statement.PrintTree(1)
}

type Statement interface {
	EmitStatement(*compiler)
	PrintTree(int)
}

type WhileStatement struct {
	condition Expression
	body      BlockStatement
}

func (ws WhileStatement) EmitStatement(c *compiler) {
	loopTo := c.chunkSize()
	ws.condition.EmitExpression(c)

	jumpFrom := c.chunkSize()
	c.emitJump(OpJumpIfFalse)

	ws.body.EmitStatement(c)

	loopFrom := c.chunkSize()
	c.emitJump(OpLoop)
	jumpTo := c.chunkSize()

	c.patchJump(loopFrom, loopTo)
	c.patchJump(jumpFrom, jumpTo)
}

func (ws WhileStatement) PrintTree(indent int) {
	fmt.Fprintln(os.Stderr, "While")
	ws.condition.PrintTree(indent + 1)

	ws.body.PrintTree(indent + 1)
}

type BlockStatement struct {
	statements []Statement
}

// todo: block scope
func (bs BlockStatement) EmitStatement(c *compiler) {
	for _, st := range bs.statements {
		st.EmitStatement(c)
	}
}

func (bs BlockStatement) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Block")

	for _, st := range bs.statements {
		st.PrintTree(indent + 1)
	}
}

type AssertStatement struct {
	value Expression
}

func (as AssertStatement) EmitStatement(c *compiler) {
	as.value.EmitExpression(c)
	c.emitByte(OpAssert)
}

func (as AssertStatement) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Assert")
	as.value.PrintTree(indent + 1)
}

type ExpressionStatement struct {
	value Expression
}

func (es ExpressionStatement) EmitStatement(c *compiler) {
	es.value.EmitExpression(c)
	c.emitByte(OpPop)
}

func (es ExpressionStatement) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Expression")
	es.value.PrintTree(indent + 1)
}

type Expression interface {
	EmitExpression(*compiler)
	PrintTree(int)
}

type Assignment struct {
	name  string
	value Expression
}

func (ass Assignment) EmitExpression(c *compiler) {
	ass.value.EmitExpression(c)

	name := c.makeConstant(value.StringVal(ass.name))
	c.emitBytes(OpSetGlobal, name)
}

func (ass Assignment) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Assign")
	printIndent(indent + 1)
	fmt.Fprintln(os.Stderr, ass.name)
	ass.value.PrintTree(indent + 1)
}

type LogicOr struct {
	value LogicAnd
	or    []LogicAnd
}

// todo: short circuit or with a jump
func (lo LogicOr) EmitExpression(c *compiler) {
	lo.value.EmitExpression(c)
	for _, la := range lo.or {
		la.EmitExpression(c)
		c.emitByte(OpOr)
	}
}

func (lo LogicOr) PrintTree(indent int) {
	if len(lo.or) == 0 {
		lo.value.PrintTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Or")
	lo.value.PrintTree(indent + 1)
	for _, or := range lo.or {
		or.PrintTree(indent + 1)
	}
}

type LogicAnd struct {
	value Comparison
	and   []Comparison
}

func (la LogicAnd) EmitExpression(c *compiler) {
	la.value.EmitExpression(c)
	for _, comp := range la.and {
		comp.EmitExpression(c)
		c.emitByte(OpAnd)
	}
}

func (la LogicAnd) PrintTree(indent int) {
	if len(la.and) == 0 {
		la.value.PrintTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, "And")
	la.value.PrintTree(indent + 1)
	for _, comp := range la.and {
		comp.PrintTree(indent + 1)
	}
}

type Comparison struct {
	term  Term
	items []ComparisonItem
}

// todo: could return second rather than pop but that would not
// be compatible with Lua
func (comp Comparison) EmitExpression(c *compiler) {
	comp.term.EmitExpression(c)
	for _, ci := range comp.items {
		ci.term.EmitExpression(c)
		switch ci.compareOp {
		case scanner.TokenEqualEqual:
			c.emitByte(OpEquals)
		case scanner.TokenLess:
			c.emitByte(OpLessThan)
		default:
			c.error(fmt.Sprint("Unknown comparator operator: ", ci.compareOp))
		}
	}
}

func (comp Comparison) PrintTree(indent int) {
	if len(comp.items) == 0 {
		comp.term.PrintTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, comp.items[0].compareOp)
	comp.term.PrintTree(indent + 1)
	Comparison{
		comp.items[0].term,
		comp.items[1:],
	}.PrintTree(indent + 1)
}

type ComparisonItem struct {
	compareOp scanner.TokenType
	term      Term
}

type Term struct {
	factor Factor
	items  []TermItem
}

func (t Term) EmitExpression(c *compiler) {
	t.factor.EmitExpression(c)
	for _, ti := range t.items {
		ti.factor.EmitExpression(c)
		switch ti.termOp {
		case scanner.TokenPlus:
			c.emitByte(OpAdd)
		case scanner.TokenMinus:
			c.emitByte(OpSubtract)
		default:
			c.error(fmt.Sprint("Unknown term operator: ", ti.termOp))
		}
	}
}

func (t Term) PrintTree(indent int) {
	if len(t.items) == 0 {
		t.factor.PrintTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, t.items[0].termOp)
	t.factor.PrintTree(indent + 1)

	Term{
		t.items[0].factor,
		t.items[1:],
	}.PrintTree(indent + 1)
}

type TermItem struct {
	termOp scanner.TokenType
	factor Factor
}

type Factor struct {
	unary Unary
	items []FactorItem
}

func (f Factor) EmitExpression(c *compiler) {
	f.unary.EmitUnary(c)
	for _, u := range f.items {
		u.unary.EmitUnary(c)
		switch u.factorOp {
		case scanner.TokenStar:
			c.emitByte(OpMult)
		case scanner.TokenSlash:
			c.emitByte(OpDivide)
		default:
			c.error(fmt.Sprint("Unkown factor operator: ", u.factorOp))
		}
	}
}

func (f Factor) PrintTree(indent int) {
	if len(f.items) == 0 {
		f.unary.PrintTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, f.items[0].factorOp)
	f.unary.PrintTree(indent + 1)

	Factor{
		f.items[0].unary,
		f.items[1:],
	}.PrintTree(indent + 1)
}

type FactorItem struct {
	factorOp scanner.TokenType
	unary    Unary
}

type Unary interface {
	EmitUnary(*compiler)
	PrintTree(int)
}

type NegateUnary struct {
	unary Unary
}

func (nu NegateUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNegate)
}

func (nu NegateUnary) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Negate")
	nu.unary.PrintTree(indent + 1)
}

type NotUnary struct {
	unary Unary
}

func (nu NotUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNot)
}

func (nu NotUnary) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Not")
	nu.unary.PrintTree(indent + 1)
}

type BaseUnary struct {
	exponent Exponent
}

func (nu BaseUnary) EmitUnary(c *compiler) {
	nu.exponent.base.EmitPrimary(c)

	if nu.exponent.exp != nil {
		nu.exponent.exp.EmitPrimary(c)
		panic("todo exponentiation")
	}
}

func (nu BaseUnary) PrintTree(indent int) {
	if nu.exponent.exp == nil {
		nu.exponent.base.PrintTree(indent)
	} else {
		printIndent(indent)
		fmt.Fprintln(os.Stderr, "Exp")
		nu.exponent.base.PrintTree(indent + 1)
		nu.exponent.exp.PrintTree(indent + 1)
	}
}

type Exponent struct {
	base Primary
	exp  Primary
}

type Primary interface {
	EmitPrimary(*compiler)
	PrintTree(int)
}

type ValuePrimary struct {
	value value.Value
}

func (vp ValuePrimary) EmitPrimary(c *compiler) {
	b := c.makeConstant(vp.value)
	c.emitBytes(OpConstant, b)
}

func (vp ValuePrimary) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, vp.value)
}

func NumberPrimary(f float64) ValuePrimary {
	return ValuePrimary{
		value: value.Number(f),
	}
}

func BooleanPrimary(b bool) ValuePrimary {
	return ValuePrimary{
		value: value.Boolean(b),
	}
}

func StringPrimary(s string) ValuePrimary {
	return ValuePrimary{
		value: value.StringVal(s),
	}
}

func NilPrimary() ValuePrimary {
	return ValuePrimary{
		value: value.Nil{},
	}
}

type GlobalPrimary string

func (gp GlobalPrimary) EmitPrimary(c *compiler) {
	constant := c.makeConstant(value.StringVal(string(gp)))
	c.emitBytes(OpGetGlobal, constant)
}

func (gp GlobalPrimary) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintf(os.Stderr, "Global/%s\n", string(gp))
}

type TableLiteral struct {
	entries []Pair
}

func (tl TableLiteral) EmitPrimary(c *compiler) {
	c.emitByte(OpCreateTable)

	for _, ent := range mapPairs(tl.entries) {
		ent.EmitPair(c)
	}

	for _, ent := range valuePairs(tl.entries) {
		ent.EmitPair(c)
	}
}

func valuePairs(pairs []Pair) []Pair {
	var mapPairs []Pair

	for _, p := range pairs {
		switch p.(type) {
		case Value:
			mapPairs = append(mapPairs, p)
		}
	}

	return mapPairs
}

func mapPairs(pairs []Pair) []Pair {
	var mapPairs []Pair

	for _, p := range pairs {
		switch p.(type) {
		case Value:
		default:
			mapPairs = append(mapPairs, p)
		}
	}

	return mapPairs
}

func (tl TableLiteral) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Table")

	for _, entry := range tl.entries {
		entry.PrintTree(indent + 1)
	}
}

type Pair interface {
	EmitPair(c *compiler)
	PrintTree(indent int)
}

type Value struct {
	value Expression
}

func (v Value) EmitPair(c *compiler) {
	v.value.EmitExpression(c)
	c.emitByte(OpInsertTable)
}

func (v Value) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Value")
	v.value.PrintTree(indent + 1)
}

type StringPair struct {
	key   Primary
	value Expression
}

func (sp StringPair) EmitPair(c *compiler) {
	sp.key.EmitPrimary(c)
	sp.value.EmitExpression(c)
	c.emitByte(OpSetTable)
}

func (sp StringPair) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "StringPair")
	printIndent(indent)
	fmt.Fprintln(os.Stderr, sp.key)
	sp.value.PrintTree(indent + 1)
}

type LiteralPair struct {
	key   Expression
	value Expression
}

func (lp LiteralPair) EmitPair(c *compiler) {
	lp.key.EmitExpression(c)
	lp.value.EmitExpression(c)
	c.emitByte(OpSetTable)
}

func (lp LiteralPair) PrintTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "LiteralPair")
	lp.key.PrintTree(indent + 1)
	lp.value.PrintTree(indent + 1)
}

func printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(os.Stderr, "  ")
	}
}
