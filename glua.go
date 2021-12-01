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
	vm := interpreter.NewVm()
	for line, _, err := reader.ReadLine(); err == nil; line, _, err = reader.ReadLine() {
		val, err := interpreter.FromString(&vm, string(line)).Interpret()

		if !err.IsEmpty() {
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

	val, intErr := interpreter.FromBufio(reader).Interpret()

	if !intErr.IsEmpty() {
		switch err.(type) {
		case scanner.ScanError:
			fmt.Println(err)
			os.Exit(1)
		case compiler.CompileError:
			fmt.Println(err)
			os.Exit(2)
		case interpreter.RuntimeError:
			fmt.Println(err)
			os.Exit(3)
		default:
			fmt.Println("Unexpected error: ", err)
			os.Exit(4)
		}
	}

	fmt.Println("Result: ", val)
}
