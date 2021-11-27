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

func TestCompileSingleExpression(t *testing.T) {
	text := "1 + 2 + 3 + 4"

	runWithoutError(t, text)
}

func TestCompileArithmeticOperations(t *testing.T) {
	text := "-1 * 2 + -3 + 4"

	runWithoutError(t, text)
}

func TestCompileLotsOfNegatives(t *testing.T) {
	text := "-1 * -1 * -1 * -1 * -1 + 1"

	runWithoutError(t, text)
}

func TestCompileAssertStatement(t *testing.T) {
	text := "assert true"

	runWithoutError(t, text)
}
