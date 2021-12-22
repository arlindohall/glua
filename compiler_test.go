package main

import (
	"arlindohall/glua/constants"
	"arlindohall/glua/interpreter"
	"fmt"
	"testing"
)

func expectNoErrors(t *testing.T, text string) {
	if constants.PrintTokens {
		fmt.Println("~~~~~~~~~~ Running program ~~~~~~~~~~")
		fmt.Println(text)
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	}

	vm := interpreter.NewVm()
	_, err := interpreter.FromString(&vm, text).Interpret()

	if !err.IsEmpty() {
		fmt.Println("Error running test:")
		fmt.Println(err)
		t.FailNow()
	}
}

func TestSingleExpression(t *testing.T) {
	text := "1 + 2 + 3 + 4"

	expectNoErrors(t, text)
}

func TestArithmeticOperations(t *testing.T) {
	text := "-1 * 2 + -3 + 4"

	expectNoErrors(t, text)
}

func TestLotsOfNegatives(t *testing.T) {
	text := "-1 * -1 * -1 * -1 * -1 + 1"

	expectNoErrors(t, text)
}

func TestAssertStatement(t *testing.T) {
	text := "assert true"

	expectNoErrors(t, text)
}

func TestAssertNotFalse(t *testing.T) {
	text := "assert !false"

	expectNoErrors(t, text)
}

func TestArithmeticAll(t *testing.T) {
	text := "assert 1 * 2 + 3 / 4 - 5 / 6 * 7 * 8 * 3 + 9 + 1 / 4 == -128"

	expectNoErrors(t, text)
}

func TestGrouping(t *testing.T) {
	text := "assert 3 * (1 + 2) * 3 == 27"

	expectNoErrors(t, text)
}

func TestAndExpression(t *testing.T) {
	text := "assert 1 and 2 == 2 or 3 == 3 == 3"

	expectNoErrors(t, text)
}

func TestStringLiteral(t *testing.T) {
	text := "assert \"abc\" == \"abc\""

	expectNoErrors(t, text)
}

func TestNilFalsey(t *testing.T) {
	text := "assert ! nil"

	expectNoErrors(t, text)
}

func TestZeroTruthy(t *testing.T) {
	text := "assert 0"

	expectNoErrors(t, text)
}

func TestEmptyStringTruthy(t *testing.T) {
	text := "assert \"\""

	expectNoErrors(t, text)
}

func TestGlobalVariable(t *testing.T) {
	text := `
	global x = 10
	assert x == 10`

	expectNoErrors(t, text)
}

func TestWhileStatement(t *testing.T) {
	text := `
	global x = 1
	while x < 10 do
		x = x + 1
	end

	assert x == 10
	`

	expectNoErrors(t, text)
}

func TestTableLiteral(t *testing.T) {
	text := "x = {1, 2, 3}"

	expectNoErrors(t, text)
}

func TestTableLiteralKeywordArgs(t *testing.T) {
	text := "x = {a=1, b=2, c=3}"

	expectNoErrors(t, text)
}

func TestTableLiteralPairArgs(t *testing.T) {
	text := "x = {[2]=1, [3]=2, [17]=3}"

	expectNoErrors(t, text)
}

func TestGetTableAttribute(t *testing.T) {
	text := `
	x = {a=1}
	assert x.a == 1
	assert x.b == nil
	`

	expectNoErrors(t, text)
}

func TestSetTableAttribute(t *testing.T) {
	text := `
	x = {}
	x.a = 1
	assert x.a == 1
	`

	expectNoErrors(t, text)
}

func TestGetTableBracketNotation(t *testing.T) {
	text := `
	x = {}
	x[4] = 1
	assert x[4] == 1`

	expectNoErrors(t, text)
}

func TestChainedAssignment(t *testing.T) {
	text := `
	x, y = true, true
	t, u = {}, {}
	t.x, u.x = true, true
	t[true], u[true] = 10

	assert x and y
	assert t and u
	assert t.x and u.x
	assert t[true] == 10 and u[true] == 10
	`

	expectNoErrors(t, text)
}

func TestLocalScope(t *testing.T) {
	text := `
	global x = 10
	do
		assert x == 10
		local x = 5
		assert x == 5
		x = 15
		assert x == 15
	end
	assert x == 10`

	expectNoErrors(t, text)
}

func TestBasicFunction(t *testing.T) {
	text := `
	function f()
		return 1
	end

	assert f() == 1
	`

	expectNoErrors(t, text)
}

func TestFunctionOneArg(t *testing.T) {
	text := `
	function f(x) return x + 1 end
	f(1)
	assert f(1) == 2
	`

	expectNoErrors(t, text)
}

func TestCloseLocal(t *testing.T) {
	text := `
	global f

	do
		local x = 10
		function g()
			return x
		end
		f = g
		assert x == 10
	end

	assert x == nil
	assert f() == 10
	`

	expectNoErrors(t, text)
}

func TestCloseFunction(t *testing.T) {
	text := `
	function f(x)
		function g()
			return x
		end
		return g
	end

	local h = f(10)
	assert h() == 10

	local j = f(20)
	assert j() == 20
	assert h() == 10
	`

	expectNoErrors(t, text)
}

func TestIfStatement(t *testing.T) {
	text := `
	if x then
		assert false
	else
		assert true
	end
	`

	expectNoErrors(t, text)
}

func TestCounterfactual(t *testing.T) {
	text := `
	if !x then
		assert true
	else
		assert false
	end
	`

	expectNoErrors(t, text)
}

func TestBuiltinTime(t *testing.T) {
	text := "time()"

	expectNoErrors(t, text)
}

func TestArity(t *testing.T) {
	text := []string{}

	add := func(s string) {
		text = append(text, s)
	}

	add(`
	function f()
		return 1, 2
	end

	x, y = f()

	assert x == 1
	assert y == 2
	`)

	add(`
	function f()
		return 1, 2
	end

	x = f()

	assert x == 1
	`)

	add(`
	function f()
		return 1, 2
	end

	x, y= 3, f()

	assert x == 3
	assert y == 1
	`)

	add(`
	function f()
		return 1, 2
	end

	x, y= 3, f()

	assert x == 3
	assert y == 1
	`)

	for _, prog := range text {
		expectNoErrors(t, prog)
	}
}

func TestFibonacciMultipleAssignment(t *testing.T) {
	text := `
	function fib(x)
		if x <= 2 then
			return 1, 1
		end

		local f1, f2 = fib(x-1)
		return f1+f2, f1
	end

	a, b = fib(10)
	assert a == 55
	assert b == 34
	assert fib(10) == 55
	`

	expectNoErrors(t, text)
}

func TestEmptyReturn(t *testing.T) {
	text := `
	function f()
		return
	end

	assert f() == nil
	`

	expectNoErrors(t, text)
}

// func TestStressFunctionCall(t *testing.T) {
// 	text := `
// 	function fib(x)
// 		if x <= 2 then
// 			return 1, 1
// 		end

// 		local f1, f2 = fib(x-1)
// 		return f1+f2, f1
// 	end

// 	fib(1000000)
// 	`
// 	expectNoErrors(t, text)
// }

// func TestStressTableAccess(t *testing.T) {
// 	text := `
// 	x = 0
// 	t = {}
// 	while x < 10000000 do
// 		function f()
// 			return x
// 		end
// 		t[x] = f
// 		x = x + 1
// 	end

// 	y = 0
// 	x = 0
// 	while x < 10000000 do
// 		y = y + t[x]()
// 		x = x + 1
// 	end
// 	`
// 	expectNoErrors(t, text)
// }
