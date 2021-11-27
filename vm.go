package main

import (
	"fmt"
	"os"
)

type VM struct {
	ip        int
	chunk     chunk
	stack     []Value
	stackSize int
	err       error
}

type op byte

func (vm *VM) Interpret(chunk chunk) (Value, error) {
	vm.chunk = chunk

	// todo: call function
	if TraceExecution {
		fmt.Println("========== <script> ==========")
	}

	val := vm.run()

	return val, vm.err
}

func (vm *VM) run() Value {
	for {
		op := vm.readByte()

		if TraceExecution {
			debugTrace(vm)
		}

		switch op {
		case OpAssert:
			val := vm.pop()
			if !val.asBoolean() {
				// exit or break to top level?
				os.Exit(5)
			}
		case OpReturn:
			return vm.pop()
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
		case OpNot:
			val := vm.pop().asBoolean()
			vm.push(boolean{!val})
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
			return nil
		}
	}
}

func (vm *VM) pop() Value {
	vm.stackSize -= 1
	val := vm.stack[vm.stackSize]
	vm.stack[vm.stackSize] = nil

	return val
}

func (vm *VM) push(val Value) {
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
	vm.err = RuntimeError{message}
}

type RuntimeError struct {
	message string
}

func (re RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error ---> %s", re.message)
}
