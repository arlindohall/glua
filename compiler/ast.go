package compiler

import (
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"fmt"
)

type Declaration interface {
	EmitDeclaration(*compiler)
	String() string
}

type StatementDeclaration struct {
	statement Statement
}

func (sd StatementDeclaration) EmitDeclaration(c *compiler) {
	sd.statement.EmitStatement(c)
}

func (sd StatementDeclaration) String() string {
	return fmt.Sprintf("Declaration(%s)", sd.statement.String())
}

type Statement interface {
	EmitStatement(*compiler)
	String() string
}

type AssertStatement struct {
	value Expression
}

func (as AssertStatement) EmitStatement(c *compiler) {
	as.value.EmitExpression(c)
	c.emitByte(OpAssert)
}

func (as AssertStatement) String() string {
	return fmt.Sprintf("Assert(%s)", as.value.String())
}

type ExpressionStatement struct {
	value Expression
}

func (es ExpressionStatement) EmitStatement(c *compiler) {
	es.value.EmitExpression(c)
	c.emitByte(OpPop)
}

func (es ExpressionStatement) String() string {
	return fmt.Sprintf("Expression(%s)", es.value.String())
}

type Expression interface {
	EmitExpression(*compiler)
	String() string
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

func (lo LogicOr) String() string {
	if lo.or != nil {
		var ors []LogicAnd
		ors = append(ors, lo.value)
		ors = append(ors, lo.or...)
		return fmt.Sprintf("Or(%s)", ors)
	} else {
		return lo.value.String()
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

func (la LogicAnd) String() string {
	if la.and != nil {
		var ands []Comparison
		ands = append(ands, la.value)
		ands = append(ands, la.and...)
		return fmt.Sprintf("And(%v)", ands)
	} else {
		return la.value.String()
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
		panic("todo Emit comparison")
	}
}

func (comp Comparison) String() string {
	if comp.items != nil {
		panic("todo display comparison")
	} else {
		return comp.term.String()
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
			panic("unreachable")
		}
	}
}

func (t Term) String() string {
	return termString(t.factor, t.items)
}

func termString(f Factor, tis []TermItem) string {
	if len(tis) == 0 {
		return f.String()
	} else {
		return fmt.Sprintf("%s(%v, %s)", tis[0].termOp, f, termString(tis[0].factor, tis[1:]))
	}
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
			panic("unreachable")
		}
	}
}

func (f Factor) String() string {
	return factorString(f.unary, f.items)
}

func factorString(u Unary, fis []FactorItem) string {
	if len(fis) == 0 {
		return u.String()
	} else {
		return fmt.Sprintf("%s(%s, %s)", fis[0].factorOp, u.String(), factorString(fis[0].unary, fis[1:]))
	}
}

type FactorItem struct {
	factorOp scanner.TokenType
	unary    Unary
}

type Unary interface {
	EmitUnary(*compiler)
	String() string
}

type NegateUnary struct {
	unary Unary
}

func (nu NegateUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNegate)
}

func (nu NegateUnary) String() string {
	return fmt.Sprintf("Negate(%s)", nu.unary.String())
}

type NotUnary struct {
	unary Unary
}

func (nu NotUnary) EmitUnary(c *compiler) {
	nu.unary.EmitUnary(c)
	c.emitByte(OpNot)
}

func (nu NotUnary) String() string {
	return fmt.Sprintf("Not(%s)", nu.unary.String())
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

func (nu BaseUnary) String() string {
	if nu.exponent.exp == nil {
		return nu.exponent.base.String()
	} else {
		return fmt.Sprintf("Exp(%v, %v)", nu.exponent.base, nu.exponent.exp)
	}
}

type Exponent struct {
	base Primary
	exp  Primary
}

type Primary interface {
	EmitPrimary(*compiler)
	String() string
}

type ValuePrimary struct {
	value value.Value
}

func (vp ValuePrimary) EmitPrimary(c *compiler) {
	b := c.makeConstant(vp.value)
	c.emitBytes(OpConstant, b)
}

func (vp ValuePrimary) String() string {
	return vp.value.String()
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
