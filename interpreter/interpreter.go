package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/glerror"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	PrintTokens, TraceExecution bool = true, true
)

// todo: Don't provide mode
// todo: Glua should include a VM
type Glua interface {
	Interpret() (value.Value, glerror.GluaErrorChain)
}

type BufioInterpreter struct {
	text *bufio.Reader
	mode compiler.ReturnMode
	vm   *VM
}

type StringInterpreter struct {
	text string
	mode compiler.ReturnMode
	vm   *VM
}

func FromString(vm *VM, text string) Glua {
	return StringInterpreter{text, compiler.ReplMode, vm}
}

func FromBufio(reader *bufio.Reader) Glua {
	vm := NewVm()
	interpreter := BufioInterpreter{reader, compiler.RunFileMode, &vm}
	return &interpreter
}

func (interp StringInterpreter) Interpret() (value.Value, glerror.GluaErrorChain) {
	reader := bufio.NewReader(strings.NewReader(string(interp.text)))
	return interp.ToBufioInterpreter(reader).Interpret()
}

func (in StringInterpreter) ToBufioInterpreter(reader *bufio.Reader) Glua {
	interp := BufioInterpreter{
		reader,
		in.mode,
		in.vm,
	}
	return &interp
}

func (interp *BufioInterpreter) Interpret() (value.Value, glerror.GluaErrorChain) {
	reader := bufio.Reader(*interp.text)

	scan := scanner.Scanner(&reader)
	tokens, err := scan.ScanTokens()

	if !err.IsEmpty() {
		return nil, err
	}

	if PrintTokens {
		scanner.DebugTokens(tokens)
	}

	function, err := compiler.Compile(tokens, interp.mode)

	if !err.IsEmpty() {
		return nil, err
	}

	// todo: use a VM struct that is re-used on Repl
	val, err := interp.vm.Interpret(function.Chunk)

	if !err.IsEmpty() {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	return val, glerror.GluaErrorChain{}
}
