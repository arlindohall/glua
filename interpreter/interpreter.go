package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/constants"
	"arlindohall/glua/glerror"
	"arlindohall/glua/scanner"
	"arlindohall/glua/value"
	"bufio"
	"strings"
)

// todo: Don't provide mode
// todo: Glua should include a VM
type Glua interface {
	Interpret() (value.Value, glerror.GluaErrorChain)
}

// todo: it's weird to pass in the vm
// instead have the interpreter be persistent and pass in only the string
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
	return StringInterpreter{text, constants.ReplMode, vm}
}

func FromBufio(reader *bufio.Reader) Glua {
	vm := NewVm()
	interpreter := BufioInterpreter{reader, constants.RunFileMode, &vm}
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

	if constants.PrintTokens {
		scanner.DebugTokens(tokens)
	}

	function, err := compiler.Compile(tokens, interp.mode)

	if !err.IsEmpty() {
		return nil, err
	}

	// todo: use a VM struct that is re-used on Repl
	val, err := interp.vm.Interpret(function)

	if !err.IsEmpty() {
		return nil, err
	}

	return val, glerror.GluaErrorChain{}
}
