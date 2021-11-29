package interpreter

import (
	"arlindohall/glua/compiler"
	"fmt"
	"os"
)

func DebugTrace(vm *VM) {
	var trace func(int, *VM)
	switch vm.previous() {
	case compiler.OpAdd, compiler.OpSubtract, compiler.OpNot, compiler.OpNegate, compiler.OpMult, compiler.OpDivide, compiler.OpNil, compiler.OpReturn, compiler.OpPop, compiler.OpAssert, compiler.OpEquals, compiler.OpAnd, compiler.OpOr:
		trace = traceInstruction
	case compiler.OpConstant, compiler.OpSetGlobal, compiler.OpGetGlobal:
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
