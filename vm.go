package main

type VM struct {
	ip    int
	chunk chunk
}

type op byte

func Run(chunk chunk) {
	vm := VM{0, chunk}
	vm.run()
}

func (vm *VM) run() {
	for {
		op := vm.current()

		if TraceExecution {
			debugTrace(vm)
		}

		switch op {
		case OpReturn:
			return
		}

		vm.ip += 1
	}
}

func (vm *VM) current() op {
	return vm.chunk.bytecode[vm.ip]
}
