package main

import (
	"fmt"
	"os"
)

func (tt TokenType) String() string {
	switch tt {
	case TokenError:
		return "TokenError"
	case TokenNumber:
		return "TokenNumber"
	case TokenEof:
		return "TokenEof"
	case TokenPlus:
		return "TokenPlus"
	case TokenMinus:
		return "TokenMinus"
	case TokenSemicolon:
		return "TokenSemicolon"
	case TokenSlash:
		return "TokenSlash"
	case TokenStar:
		return "TokenStar"
	default:
		panic("Unrecognized TokenType")
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v/\"%v\"", t._type, t.text)
}

func (op op) String() string {
	switch op {
	case OpNil:
		return "OpNil"
	case OpConstant:
		return "OpConstant"
	case OpPop:
		return "OpPop"
	case OpReturn:
		return "OpReturn"
	case OpAdd:
		return "OpAdd"
	case OpSubtract:
		return "OpSubtract"
	case OpMult:
		return "OpMult"
	case OpDivide:
		return "OpDivide"
	default:
		panic(fmt.Sprint("Unrecognized: ", byte(op)))
	}
}

func debugTrace(vm *VM) {
	var trace func(int, *VM)
	switch vm.previous() {
	case OpAdd, OpSubtract, OpMult, OpDivide, OpNil, OpReturn, OpPop:
		trace = traceInstruction
	case OpConstant:
		trace = traceConstant
	default:
		panic(fmt.Sprint("Do not know how to trace: ", vm.previous()))
	}

	trace(vm.ip, vm)
}

func traceInstruction(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-12v      %v\n", i, vm.previous(), vm.stack[:vm.stackSize])
}

func traceConstant(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-12v %-4d %v\n", i, vm.previous(), vm.current(), vm.stack[:vm.stackSize])
}

func debugPrint(function Function) {
	bytecode := function.chunk.bytecode

	if function.name == "" {
		fmt.Fprintln(os.Stderr, "---------- <script> ----------")
	} else {
		fmt.Fprintln(os.Stderr, "----------", function.name, "----------")
	}

	i := 0
	var print func(int, []op) int
	for i < len(bytecode) {
		switch bytecode[i] {
		case OpConstant:
			print = printConstant
		case OpAdd, OpSubtract, OpMult, OpDivide, OpNil, OpReturn, OpPop:
			print = printInstruction
		default:
			panic(fmt.Sprint("Unknown op for debug print: ", bytecode[i]))
		}
		i = print(i, bytecode)
	}

	fmt.Fprintln(os.Stderr)
}

func printInstruction(i int, bytecode []op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-12v\n", i, bytecode[i])
	return i + 1
}

func printConstant(i int, bytecode []op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-12v %-4d\n", i, bytecode[i], bytecode[i+1])
	return i + 2
}

func debugTokens(tokens []Token) {
	for _, token := range tokens {
		if token._type == TokenSemicolon {
			fmt.Fprintln(os.Stderr, ";")
			continue
		}
		fmt.Fprint(os.Stderr, token, " ")
	}

	fmt.Fprintln(os.Stderr)
}
