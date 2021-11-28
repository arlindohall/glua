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

type StatementDeclaration struct {
	statement Statement
}

func (sd StatementDeclaration) EmitDeclaration(c *compiler) {
	sd.statement.EmitStatement(c)
}

func (sd StatementDeclaration) PrintTree() {
	fmt.Fprintln(os.Stderr, "Statement")
	sd.statement.printTree(1)
}

type Statement interface {
	EmitStatement(*compiler)
	printTree(int)
}

type AssertStatement struct {
	value Expression
}

func (as AssertStatement) EmitStatement(c *compiler) {
	as.value.EmitExpression(c)
	c.emitByte(OpAssert)
}

func (as AssertStatement) printTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Assert")
	as.value.printTree(indent + 1)
}

type ExpressionStatement struct {
	value Expression
}

func (es ExpressionStatement) EmitStatement(c *compiler) {
	es.value.EmitExpression(c)
	c.emitByte(OpPop)
}

func (es ExpressionStatement) printTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Expression")
	es.value.printTree(indent + 1)
}

type Expression interface {
	EmitExpression(*compiler)
	printTree(int)
}

type LogicOr struct {
	value LogicAnd
	or    []LogicAnd
}

func (lo LogicOr) EmitExpression(c *compiler) {
	lo.value.EmitExpression(c)
	for _, la := range lo.or {
		la.EmitExpression(c)
		panic("todo logic or")
	}
}

func (lo LogicOr) printTree(indent int) {
	if len(lo.or) == 0 {
		lo.value.printTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Or")
	lo.value.printTree(indent + 1)
	for _, or := range lo.or {
		or.printTree(indent + 1)
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
		panic("todo logic and")
	}
}

func (la LogicAnd) printTree(indent int) {
	if len(la.and) == 0 {
		la.value.printTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprint(os.Stderr, "And")
	la.value.printTree(indent + 1)
	for _, comp := range la.and {
		comp.printTree(indent + 1)
	}
}

type Comparison struct {
	term  Term
	items []ComparisonItem
}

func (comp Comparison) EmitExpression(c *compiler) {
	comp.term.EmitExpression(c)
	for _, ci := range comp.items {
		ci.term.EmitExpression(c)
		switch ci.compareOp {
		case scanner.TokenEqualEqual:
			c.emitByte(OpEquals)
		default:
			c.error(fmt.Sprint("Unknown comparator operator: ", ci.compareOp))
		}
	}
}

func (comp Comparison) printTree(indent int) {
	if len(comp.items) == 0 {
		comp.term.printTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Compare")
	comp.term.printTree(indent + 1)
	for _, comp := range comp.items {
		printIndent(indent + 1)
		fmt.Println(comp.compareOp)
		comp.term.printTree(indent + 1)
	}
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

func (t Term) printTree(indent int) {
	if len(t.items) == 0 {
		t.factor.printTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, t.items[0].termOp)
	t.factor.printTree(indent + 1)

	Term{
		t.items[0].factor,
		t.items[1:],
	}.printTree(indent + 1)
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

func (f Factor) printTree(indent int) {
	if len(f.items) == 0 {
		f.unary.printTree(indent)
		return
	}

	printIndent(indent)
	fmt.Fprintln(os.Stderr, f.items[0].factorOp)
	f.unary.printTree(indent + 1)

	Factor{
		f.items[0].unary,
		f.items[1:],
	}.printTree(indent + 1)
}

type FactorItem struct {
	factorOp scanner.TokenType
	unary    Unary
}

type Unary interface {
	EmitUnary(*compiler)
	printTree(int)
}

type NegateUnary struct {
	unary Unary
}

func (nu NegateUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNegate)
}

func (nu NegateUnary) printTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Negate")
	nu.unary.printTree(indent + 1)
}

type NotUnary struct {
	unary Unary
}

func (nu NotUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNot)
}

func (nu NotUnary) printTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, "Not")
	nu.unary.printTree(indent + 1)
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

func (nu BaseUnary) printTree(indent int) {
	if nu.exponent.exp == nil {
		nu.exponent.base.printTree(indent)
	} else {
		printIndent(indent)
		fmt.Fprintln(os.Stderr, "Exp")
		nu.exponent.base.printTree(indent + 1)
		nu.exponent.exp.printTree(indent + 1)
	}
}

type Exponent struct {
	base Primary
	exp  Primary
}

type Primary interface {
	EmitPrimary(*compiler)
	printTree(int)
}

type ValuePrimary struct {
	value value.Value
}

func (vp ValuePrimary) EmitPrimary(c *compiler) {
	b := c.makeConstant(vp.value)
	c.emitBytes(OpConstant, b)
}

func (vp ValuePrimary) printTree(indent int) {
	printIndent(indent)
	fmt.Fprintln(os.Stderr, vp.value)
}

func NumberPrimary(f float64) ValuePrimary {
	return ValuePrimary{
		value: value.Number{Val: f},
	}
}

func BooleanPrimary(b bool) ValuePrimary {
	return ValuePrimary{
		value: value.Boolean{Val: b},
	}
}

// type StringPrimary struct {
// 	value string
// }

// type IdentifierPrimary struct {
// 	value string
// }

func printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(os.Stderr, "  ")
	}
}
