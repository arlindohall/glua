package main

import "testing"

func TestCompileSingleExpression(t *testing.T) {
	text := "1 + 2 + 3 + 4"

	fromString(text).Interpret()
}

func TestCompileArithmeticOperations(t *testing.T) {
	text := "-1 * 2 + -3 + 4"

	fromString(text).Interpret()
}

func TestCompileLotsOfNegatives(t *testing.T) {
	text := "-1 * -1 * -1 * -1 * -1 + 1"

	fromString(text).Interpret()
}
