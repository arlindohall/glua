package compiler

import (
	"fmt"
	"os"
)

func ByteName(op byte) string {
	switch op {
	case OpAssert:
		return "OpAssert"
	case OpAssignStart:
		return "OpAssignStart"
	case OpAssignCleanup:
		return "OpAssignCleanup"
	case OpNil:
		return "OpNil"
	case OpZero:
		return "OpZero"
	case OpLocalAllocate:
		return "OpLocalAllocate"
	case OpLocalCleanup:
		return "OpLocalCleanup"
	case OpGetGlobal:
		return "OpGetGlobal"
	case OpSetGlobal:
		return "OpSetGlobal"
	case OpGetLocal:
		return "OpGetLocal"
	case OpSetLocal:
		return "OpSetLocal"
	case OpCreateTable:
		return "OpCreateTable"
	case OpSetTable:
		return "OpSetTable"
	case OpInitTable:
		return "OpInitTable"
	case OpGetTable:
		return "OpGetTable"
	case OpInsertTable:
		return "OpInsertTable"
	case OpClosure:
		return "OpClosure"
	case OpCreateUpvalue:
		return "OpCreateUpvalue"
	case OpCloseUpvalues:
		return "OpCloseUpvalues"
	case OpGetUpvalue:
		return "OpGetUpvalue"
	case OpSetUpvalue:
		return "OpSetUpvalue"
	case OpCall:
		return "OpCall"
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
	case OpLess:
		return "OpLess"
	case OpGreater:
		return "OpGreater"
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
	var print func(int, []byte) int
	for i < len(bytecode) {
		switch bytecode[i] {
		case OpConstant, OpSetGlobal, OpGetGlobal, OpSetLocal, OpGetLocal,
			OpCloseUpvalues, OpGetUpvalue, OpSetUpvalue, OpReturn:
			print = printConstant
		case OpAdd, OpSubtract, OpNot, OpNegate, OpMult, OpDivide, OpNil,
			OpPop, OpAssert, OpEquals, OpLess, OpGreater, OpAnd, OpOr,
			OpCreateTable, OpSetTable, OpInsertTable, OpInitTable, OpGetTable, OpZero,
			OpClosure, OpAssignStart, OpAssignCleanup, OpLocalAllocate, OpLocalCleanup:
			print = printInstruction
		case OpCreateUpvalue:
			print = printUpvalue
		case OpCall:
			print = printCall
		case OpLoop:
			print = printLoop
		case OpJumpIfFalse:
			print = printJump
		default:
			panic(fmt.Sprint("Unknown op for debug print: ", ByteName(bytecode[i])))
		}
		i = print(i, bytecode)
	}

	fmt.Fprintln(os.Stderr)
}

func printInstruction(i int, bytecode []byte) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16v\n", i, ByteName(bytecode[i]))
	return i + 1
}

func printConstant(i int, bytecode []byte) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16s %-4d\n", i, ByteName(bytecode[i]), bytecode[i+1])
	return i + 2
}

func printCall(i int, bytecode []byte) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16s %-4d%-4v\n", i, ByteName(bytecode[i]), bytecode[i+1], bytecode[i+2] == 1)
	return i + 3
}

func printUpvalue(i int, bytecode []byte) int {
	fmt.Fprintf(os.Stderr, "%04d | %-16s %-4d%-4v\n", i, ByteName(bytecode[i]), bytecode[i+1], bytecode[i+2] == 1)
	return i + 3
}

func printJump(i int, bytecode []byte) int {
	jump := MergeBytes(byte(bytecode[i+1]), byte(bytecode[i+2]))
	start := i + 3
	to := start + jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d)\n", i, ByteName(bytecode[i]), jump, start, to)
	return i + 3
}

func printLoop(i int, bytecode []byte) int {
	jump := MergeBytes(byte(bytecode[i+1]), byte(bytecode[i+2]))
	start := i + 3
	to := start - jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d)\n", i, ByteName(bytecode[i]), jump, start, to)
	return i + 3
}
