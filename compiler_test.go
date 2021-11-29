package main

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/interpreter"
	"fmt"
	"testing"
)

func runWithoutError(t *testing.T, text string) {
	_, err := interpreter.FromString(text).Interpret(compiler.RunFileMode)

	if err != nil {
		fmt.Println("Error running test: ", err)
		t.FailNow()
	}
}

func TestSingleExpression(t *testing.T) {
	text := "1 + 2 + 3 + 4"

	runWithoutError(t, text)
}

func TestArithmeticOperations(t *testing.T) {
	text := "-1 * 2 + -3 + 4"

	runWithoutError(t, text)
}

func TestLotsOfNegatives(t *testing.T) {
	text := "-1 * -1 * -1 * -1 * -1 + 1"

	runWithoutError(t, text)
}

func TestAssertStatement(t *testing.T) {
	text := "assert true"

	runWithoutError(t, text)
}

func TestAssertNotFalse(t *testing.T) {
	text := "assert !false"

	runWithoutError(t, text)
}

func TestArithmeticAll(t *testing.T) {
	text := "assert 1 * 2 + 3 / 4 - 5 / 6 * 7 * 8 * 3 + 9 + 1 / 4 == -128"

	runWithoutError(t, text)
}

func TestAndExpression(t *testing.T) {
	text := "assert 1 and 2 == 2 or 3 == 3 == 3"

	runWithoutError(t, text)
}

func TestStringLiteral(t *testing.T) {
	text := "assert \"abc\" == \"abc\""

	runWithoutError(t, text)
}

func TestNilFalsey(t *testing.T) {
	text := "assert ! nil"

	runWithoutError(t, text)
}

func TestZeroTruthy(t *testing.T) {
	text := "assert 0"

	runWithoutError(t, text)
}

func TestEmptyStringTruthy(t *testing.T) {
	text := "assert \"\""

	runWithoutError(t, text)
}

func TestGlobalVariable(t *testing.T) {
	text := `global x = 10
	assert x == 10`

	runWithoutError(t, text)
}
