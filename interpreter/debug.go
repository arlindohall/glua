package interpreter

import (
	"arlindohall/glua/compiler"
	"fmt"
	"os"
)

func DebugTrace(vm *VM) {
	var trace func(int, *VM)
	switch vm.previous() {
	case compiler.OpConstant, compiler.OpSetGlobal, compiler.OpGetGlobal, compiler.OpSetLocal, compiler.OpGetLocal,
		compiler.OpCloseUpvalues, compiler.OpGetUpvalue, compiler.OpSetUpvalue:
		trace = traceConstant
	case compiler.OpAdd, compiler.OpSubtract, compiler.OpNot, compiler.OpNegate, compiler.OpMult, compiler.OpDivide, compiler.OpNil,
		compiler.OpReturn, compiler.OpPop, compiler.OpAssert, compiler.OpLessThan, compiler.OpEquals, compiler.OpAnd, compiler.OpOr,
		compiler.OpCreateTable, compiler.OpSetTable, compiler.OpInsertTable, compiler.OpInitTable, compiler.OpGetTable, compiler.OpZero,
		compiler.OpCall, compiler.OpClosure:
		trace = traceInstruction
	case compiler.OpCreateUpvalue:
		trace = traceUpvalue
	case compiler.OpJumpIfFalse:
		trace = traceJump
	case compiler.OpLoop:
		trace = traceLoop
	default:
		panic(fmt.Sprint("Do not know how to trace: ", compiler.ByteName(vm.previous())))
	}

	trace(vm.frame.ip, vm)
}

func traceInstruction(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-16v                           %v\n", i, compiler.ByteName(vm.previous()), vm.stack[:vm.stackSize])
}

func traceConstant(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-4d                      %v\n", i, compiler.ByteName(vm.previous()), vm.current(), vm.stack[:vm.stackSize])
}

func traceUpvalue(i int, vm *VM) {
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-4d%-4v                  %v\n", i, compiler.ByteName(vm.previous()), vm.current(), vm.next() == 1, vm.stack[:vm.stackSize])
}

func traceJump(i int, vm *VM) {
	jump := compiler.MergeBytes(byte(vm.current()), byte(vm.next()))
	start := i + 3
	to := start + jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d) %v\n", i, compiler.ByteName(vm.previous()), jump, start, to, vm.stack[:vm.stackSize])
}

func traceLoop(i int, vm *VM) {
	jump := compiler.MergeBytes(byte(vm.current()), byte(vm.next()))
	start := i + 3
	to := start - jump
	fmt.Fprintf(os.Stderr, "%04d | %-16v %-6d (%-6d -> %-6d) %v\n", i, compiler.ByteName(vm.previous()), jump, start, to, vm.stack[:vm.stackSize])
}
