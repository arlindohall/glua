package main

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/interpreter"
	"fmt"
	"testing"
)

func expectNoErrors(t *testing.T, text string) {
	_, err := interpreter.FromString(text).Interpret(compiler.RunFileMode)

	if !err.IsEmpty() {
		fmt.Println("Error running test: ", err)
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
	text := `global x = 10
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
