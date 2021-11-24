package main

import "fmt"

type VM struct {
	ip       int
	bytecode []op
}

type op byte

func Run(bytecode []op) {
	vm := VM{0, bytecode}
	vm.run()
}

func (vm *VM) run() {
	for {
		op := vm.bytecode[vm.ip]
		vm.ip += 1

		if TraceExecution {
			fmt.Println("Running op=", op)
		}

		switch op {
		case OpReturn:
			return
		}
	}
}
