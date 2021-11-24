package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	PrintTokens, PrintBytecode, TraceExecution bool = true, true, true
)

func main() {
	if len(os.Args) <= 1 {
		// todo: Use repl
		fmt.Println("Running REPL...")
		return
	}

	fileName := os.Args[1]

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file", fileName, err)
		return
	}

	reader := bufio.NewReader(file)

	scanner := Scanner(reader)
	tokens, err := scanner.ScanTokens()

	if err != nil {
		fmt.Println(err)
	}

	if PrintTokens {
		debugTokens(tokens)
	}

	function := Compile(tokens)

	if PrintBytecode {
		debugPrint(function)
	}

	// todo: use a VM struct that is re-used on Repl
	Interpret(function.chunk)
}
