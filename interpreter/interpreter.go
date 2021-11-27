package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	PrintTokens, PrintBytecode, TraceExecution bool = true, true, true
)

// todo: Interpret should return a value for printing
// todo: Don't provide mode
type Glua interface {
	Interpret(mode compiler.ReturnMode) (value.Value, error)
}

type BufioInterpreter bufio.Reader

type StringInterpreter string

func FromString(text string) Glua {
	return StringInterpreter(text)
}

func FromBufio(reader *bufio.Reader) Glua {
	interpreter := BufioInterpreter(*reader)
	return &interpreter
}

func (text StringInterpreter) Interpret(mode compiler.ReturnMode) (value.Value, error) {
	reader := bufio.NewReader(strings.NewReader(string(text)))

	return FromBufio(reader).Interpret(mode)
}

func (text *BufioInterpreter) Interpret(mode compiler.ReturnMode) (value.Value, error) {
	reader := bufio.Reader(*text)

	scan := scanner.Scanner(&reader)
	tokens, err := scan.ScanTokens()

	if err != nil {
		return nil, err
	}

	if PrintTokens {
		scanner.DebugTokens(tokens)
	}

	function, err := compiler.Compile(tokens, mode)

	if err != nil {
		return nil, err
	}

	if PrintBytecode {
		compiler.DebugPrint(function)
	}

	vm := VM{}
	// todo: use a VM struct that is re-used on Repl
	val, err := vm.Interpret(function.Chunk)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	return val, nil
}
