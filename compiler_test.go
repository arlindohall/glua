package main

import (
	"fmt"
	"testing"
)

func runWithoutError(t *testing.T, text string) {
	_, err := fromString(text).Interpret(RunFileMode)

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
