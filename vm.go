package main

import "fmt"

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
			val2 := vm.pop()
			val1 := vm.pop()

			switch {
			case val1.isNumber() && val2.isNumber():
				vm.push(number{val1.asNumber() - val2.asNumber()})
			default:
				panic("Cannot subtract two non-numbers")
			}
		case OpAdd:
			val2 := vm.pop()
			val1 := vm.pop()

			switch {
			case val1.isNumber() && val2.isNumber():
				vm.push(number{val1.asNumber() + val2.asNumber()})
			default:
				panic("Cannot add two non-numbers")
			}
		default:
			panic(fmt.Sprint("Do not know how to perform: ", op))
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
