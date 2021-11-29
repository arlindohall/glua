package interpreter

import (
	"arlindohall/glua/compiler"
	"fmt"
	"os"
)

func DebugTrace(vm *VM) {
	var trace func(int, *VM)
	switch vm.previous() {
	case compiler.OpAdd, compiler.OpSubtract, compiler.OpNot, compiler.OpNegate, compiler.OpMult, compiler.OpDivide, compiler.OpNil, compiler.OpReturn, compiler.OpPop, compiler.OpAssert, compiler.OpLessThan, compiler.OpEquals, compiler.OpAnd, compiler.OpOr:
		trace = traceInstruction
	case compiler.OpConstant, compiler.OpSetGlobal, compiler.OpGetGlobal:
		trace = traceConstant
	case compiler.OpJumpIfFalse:
		trace = traceJump
	case compiler.OpLoop:
		trace = traceLoop
	default:
		panic(fmt.Sprint("Do not know how to trace: ", vm.previous()))
	}

	trace(vm.ip, vm)
}

func traceInstruction(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-16v                           %v\n", i, vm.previous(), vm.stack[:vm.stackSize])
}

func traceConstant(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-4d                      %v\n", i, vm.previous(), vm.current(), vm.stack[:vm.stackSize])
}

func traceJump(i int, vm *VM) {
	jump := compiler.MergeBytes(byte(vm.current()), byte(vm.next()))
	start := i + 3
	to := start + jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d) %v\n", i, vm.previous(), jump, start, to, vm.stack[:vm.stackSize])
}

func traceLoop(i int, vm *VM) {
	jump := compiler.MergeBytes(byte(vm.current()), byte(vm.next()))
	start := i + 3
	to := start - jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d) %v\n", i, vm.previous(), jump, start, to, vm.stack[:vm.stackSize])
}
