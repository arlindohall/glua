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
	case OpZero:
		return "OpZero"
	case OpGetGlobal:
		return "OpGetGlobal"
	case OpSetGlobal:
		return "OpSetGlobal"
	case OpCreateTable:
		return "OpCreateTable"
	case OpSetTable:
		return "OpSetTable"
	case OpGetTable:
		return "OpGetTable"
	case OpInsertTable:
		return "OpInsertTable"
	case OpConstant:
		return "OpConstant"
	case OpPop:
		return "OpPop"
	case OpReturn:
		return "OpReturn"
	case OpJumpIfFalse:
		return "OpJumpIfFalse"
	case OpLoop:
		return "OpLoop"
	case OpEquals:
		return "OpEquals"
	case OpLessThan:
		return "OpLessThan"
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
	case OpAnd:
		return "OpAnd"
	case OpOr:
		return "OpOr"
	default:
		panic(fmt.Sprint("Unrecognized Stringer for op: ", byte(op)))
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
		case OpConstant, OpSetGlobal, OpGetGlobal:
			print = printConstant
		case OpAdd, OpSubtract, OpNot, OpNegate, OpMult, OpDivide, OpNil,
			OpReturn, OpPop, OpAssert, OpEquals, OpLessThan, OpAnd, OpOr,
			OpCreateTable, OpSetTable, OpInsertTable, OpGetTable, OpZero:
			print = printInstruction
		case OpLoop:
			print = printLoop
		case OpJumpIfFalse:
			print = printJump
		default:
			panic(fmt.Sprint("Unknown op for debug print: ", bytecode[i]))
		}
		i = print(i, bytecode)
	}

	fmt.Fprintln(os.Stderr)
}

func printInstruction(i int, bytecode []Op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16v\n", i, bytecode[i])
	return i + 1
}

func printConstant(i int, bytecode []Op) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-4d\n", i, bytecode[i], bytecode[i+1])
	return i + 2
}

func printJump(i int, bytecode []Op) int {
	jump := MergeBytes(byte(bytecode[i+1]), byte(bytecode[i+2]))
	start := i + 3
	to := start + jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d)\n", i, bytecode[i], jump, start, to)
	return i + 3
}

func printLoop(i int, bytecode []Op) int {
	jump := MergeBytes(byte(bytecode[i+1]), byte(bytecode[i+2]))
	start := i + 3
	to := start - jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d)\n", i, bytecode[i], jump, start, to)
	return i + 3
}
