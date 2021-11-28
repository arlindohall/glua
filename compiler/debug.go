package compiler

import (
	"fmt"
	"os"
)

func (op Op) String() string {
	switch op {
	case OpAssert:
		return "OpAssert"
	case OpNil:
		return "OpNil"
	case OpConstant:
		return "OpConstant"
	case OpPop:
		return "OpPop"
	case OpReturn:
		return "OpReturn"
	case OpEquals:
		return "OpEquals"
	case OpAdd:
		return "OpAdd"
	case OpSubtract:
		return "OpSubtract"
	case OpNegate:
		return "OpNegate"
	case OpNot:
		return "OpNot"
	case OpMult:
		return "OpMult"
	case OpDivide:
		return "OpDivide"
	default:
		panic(fmt.Sprint("Unrecognized: ", byte(op)))
	}
}

func DebugPrint(function Function) {
	bytecode := function.Chunk.Bytecode

	if function.Name == "" {
		fmt.Fprintln(os.Stderr, "---------- <script> ----------")
	} else {
		fmt.Fprintln(os.Stderr, "----------", function.Name, "----------")
	}

	i := 0
	var print func(int, []Op) int
	for i < len(bytecode) {
		switch bytecode[i] {
		case OpConstant:
			print = printConstant
		case OpAdd, OpSubtract, OpNot, OpNegate, OpMult, OpDivide, OpNil, OpReturn, OpPop, OpAssert, OpEquals:
			print = printInstruction
		default:
			panic(fmt.Sprint("Unknown op for debug print: ", bytecode[i]))
		}
		i = print(i, bytecode)
	}

	fmt.Fprintln(os.Stderr)
}

func printInstruction(i int, bytecode []Op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-12v\n", i, bytecode[i])
	return i + 1
}

func printConstant(i int, bytecode []Op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-12v %-4d\n", i, bytecode[i], bytecode[i+1])
	return i + 2
}
