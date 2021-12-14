package main

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/interpreter"
	"arlindohall/glua/scanner"
	"bufio"
	"fmt"
	"os"
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

	// todo: add history
	// todo: read a whole declaration at a time, not a line
	vm := interpreter.NewVm()
	for line, _, err := reader.ReadLine(); err == nil; line, _, err = reader.ReadLine() {
		val, err := interpreter.FromString(&vm, string(line)).Interpret()

		if !err.IsEmpty() {
			fmt.Println(err)
			vm.ClearErrors()
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

	val, intErrs := interpreter.FromBufio(reader).Interpret()

	if !intErrs.IsEmpty() {
		fmt.Fprintln(os.Stderr, err)
		switch intErrs.First().(type) {
		case scanner.ScanError:
			os.Exit(1)
		case compiler.CompileError:
			os.Exit(2)
		case interpreter.RuntimeError:
			os.Exit(3)
		default:
			os.Exit(4)
		}
	}

	fmt.Println("Result: ", val)
}
