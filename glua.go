package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	PrintTokens, PrintBytecode, TraceExecution bool = true, true, true
)

func main() {
	if len(os.Args) <= 1 {
		// todo: Use repl
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Running REPL...")
		fmt.Print("> ")
		for line, _, err := reader.ReadLine(); err == nil; line, _, err = reader.ReadLine() {
			fmt.Print("> ")
			fromString(string(line)).Interpret()
		}
		return
	}

	fileName := os.Args[1]

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file", fileName, err)
		return
	}

	reader := bufio.NewReader(file)

	fromBufio(reader).Interpret()
}

// todo: Interpret should return a value for printing
type Glua interface {
	Interpret()
}

type BufioInterpreter bufio.Reader

type StringInterpreter string

func fromString(text string) Glua {
	return StringInterpreter(text)
}

func fromBufio(reader *bufio.Reader) Glua {
	interpreter := BufioInterpreter(*reader)
	return &interpreter
}

func (text StringInterpreter) Interpret() {
	reader := bufio.NewReader(strings.NewReader(string(text)))

	fromBufio(reader).Interpret()
}

func (text *BufioInterpreter) Interpret() {
	reader := bufio.Reader(*text)

	scanner := Scanner(&reader)
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
