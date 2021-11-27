package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	PrintTokens, PrintBytecode, TraceExecution bool = true, true, true

	RunFileMode = iota
	ReplMode
)

func main() {
	if len(os.Args) <= 1 {
		repl()
		return
	}

	fileName := os.Args[1]

	runFile(fileName)
}

func repl() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Running REPL...")
	fmt.Print("> ")

	for line, _, err := reader.ReadLine(); err == nil; line, _, err = reader.ReadLine() {
		val, err := fromString(string(line)).Interpret(ReplMode)

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(val)
		}

		fmt.Print("> ")
	}
}

func runFile(fileName string) {

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file", fileName, err)
		return
	}

	reader := bufio.NewReader(file)

	val, err := fromBufio(reader).Interpret(RunFileMode)

	if err != nil {
		switch err.(type) {
		case ScanError:
			fmt.Println(err)
			os.Exit(1)
		case CompileError:
			fmt.Println(err)
			os.Exit(2)
		case RuntimeError:
			fmt.Println(err)
			os.Exit(3)
		default:
			fmt.Println("Unexpected error: ", err)
			os.Exit(4)
		}
	}

	fmt.Println("Result: ", val)
}

type ReturnMode int

// todo: Interpret should return a value for printing
// todo: Don't provide mode
type Glua interface {
	Interpret(mode ReturnMode) (Value, error)
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

func (text StringInterpreter) Interpret(mode ReturnMode) (Value, error) {
	reader := bufio.NewReader(strings.NewReader(string(text)))

	return fromBufio(reader).Interpret(mode)
}

func (text *BufioInterpreter) Interpret(mode ReturnMode) (Value, error) {
	reader := bufio.Reader(*text)

	scanner := Scanner(&reader)
	tokens, err := scanner.ScanTokens()

	if err != nil {
		return nil, err
	}

	if PrintTokens {
		debugTokens(tokens)
	}

	function, err := Compile(tokens, mode)

	if err != nil {
		return nil, err
	}

	if PrintBytecode {
		debugPrint(function)
	}

	vm := VM{}
	// todo: use a VM struct that is re-used on Repl
	val, err := vm.Interpret(function.chunk)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	return val, nil
}
