package main

import (
	"fmt"
	"os"
)

type VM struct {
	ip        int
	chunk     chunk
	stack     []value
	stackSize int
}

type op byte

func Interpret(chunk chunk) {
	vm := VM{0, chunk, nil, 0}
	// todo: call function
	if TraceExecution {
		fmt.Println("========== <script> ==========")
	}
	vm.run()
}

func (vm *VM) run() {
	for {
		op := vm.readByte()

		if TraceExecution {
			debugTrace(vm)
		}

		switch op {
		case OpReturn:
			return
		case OpPop:
			vm.pop()
		case OpConstant:
			c := byte(vm.readByte())
			val := vm.chunk.constants[c]
			vm.push(val)
		case OpNil:
			vm.push(nilVal{})
		case OpSubtract:
			vm.arithmetic("subtract", func(a, b float64) float64 { return a - b })
		case OpDivide:
			vm.arithmetic("divide", func(a, b float64) float64 { return a / b })
		case OpMult:
			vm.arithmetic("multiply", func(a, b float64) float64 { return a * b })
		case OpNegate:
			val := vm.pop().asNumber()
			vm.push(number{-val})
		case OpAdd:
			val2 := vm.pop()
			val1 := vm.pop()

			switch {
			case val1.isNumber() && val2.isNumber():
				vm.push(number{val1.asNumber() + val2.asNumber()})
			default:
				vm.error("Cannot add two non-numbers")
			}
		default:
			vm.error(fmt.Sprint("Do not know how to perform: ", op))
		}
	}
}

func (vm *VM) pop() value {
	vm.stackSize -= 1
	val := vm.stack[vm.stackSize]
	vm.stack[vm.stackSize] = nil

	return val
}

func (vm *VM) push(val value) {
	if vm.stackSize >= len(vm.stack) {
		vm.stack = append(vm.stack, val)
	} else {
		vm.stack[vm.stackSize] = val
	}

	vm.stackSize += 1
}

func (vm *VM) arithmetic(name string, op func(float64, float64) float64) {
	val2 := vm.pop()
	val1 := vm.pop()

	switch {
	case val1.isNumber() && val2.isNumber():
		vm.push(number{op(val1.asNumber(), val2.asNumber())})
	default:
		vm.error(fmt.Sprintf("Cannot %s two non-numbers", name))
	}
}

func (vm *VM) readByte() op {
	c := vm.current()
	vm.advance()
	return c
}

func (vm *VM) advance() {
	vm.ip += 1
}

func (vm *VM) previous() op {
	return vm.chunk.bytecode[vm.ip-1]
}

func (vm *VM) current() op {
	return vm.chunk.bytecode[vm.ip]
}

// todo: error handling
func (vm *VM) error(message string) {
	fmt.Fprintf(os.Stdout, "Runtime error ----> %s", message)
	os.Exit(3)
}
