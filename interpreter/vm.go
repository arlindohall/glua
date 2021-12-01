package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/glerror"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

type VM struct {
	ip        int
	chunk     compiler.Chunk
	stack     []value.Value
	stackSize int
	globals   map[string]value.Value
	err       glerror.GluaErrorChain
}

func NewVm() VM {
	return VM{
		0,
		compiler.Chunk{},
		nil,
		0,
		make(map[string]value.Value),
		glerror.GluaErrorChain{},
	}
}

func (vm *VM) Interpret(chunk compiler.Chunk) (value.Value, glerror.GluaErrorChain) {
	vm.chunk = chunk

	// todo: call function
	if TraceExecution {
		fmt.Println("========== <script> ==========")
	}

	val := vm.run()

	return val, vm.err
}

func (vm *VM) run() value.Value {
	for {
		op := vm.readByte()

		if TraceExecution {
			DebugTrace(vm)
		}

		switch op {
		case compiler.OpAssert:
			val := vm.pop()
			if !val.AsBoolean() {
				// exit or break to top level?
				os.Exit(5)
			}
		case compiler.OpReturn:
			return vm.pop()
		case compiler.OpPop:
			vm.pop()
		case compiler.OpConstant:
			c := byte(vm.readByte())
			val := vm.chunk.Constants[c]
			vm.push(val)
		case compiler.OpNil:
			vm.push(value.Nil{})
		case compiler.OpZero:
			vm.push(value.Number(0))
		case compiler.OpLessThan:
			vm.compare(func(v1, v2 float64) bool { return v1 < v2 })
		case compiler.OpEquals:
			val2 := vm.pop()
			val1 := vm.pop()

			if val1.IsNumber() && val2.IsNumber() {
				vm.push(value.Boolean(val1.AsNumber() == val2.AsNumber()))
			} else if val1.IsBoolean() && val2.IsBoolean() {
				vm.push(value.Boolean(val1.AsBoolean() == val2.AsBoolean()))
			} else if val1.IsString() && val2.IsString() {
				vm.push(value.Boolean(val1.String() == val2.String()))
			} else if val1.IsNil() && val2.IsNil() {
				vm.push(value.Boolean(true))
			} else {
				vm.push(value.Boolean(false))
			}
		case compiler.OpAnd:
			val2 := vm.pop()
			val1 := vm.pop()

			vm.push(value.Boolean(val1.AsBoolean() && val2.AsBoolean()))
		case compiler.OpOr:
			val2 := vm.pop()
			val1 := vm.pop()

			vm.push(value.Boolean(val1.AsBoolean() || val2.AsBoolean()))
		case compiler.OpSubtract:
			vm.arithmetic("subtract", func(a, b float64) float64 { return a - b })
		case compiler.OpDivide:
			vm.arithmetic("divide", func(a, b float64) float64 { return a / b })
		case compiler.OpMult:
			vm.arithmetic("multiply", func(a, b float64) float64 { return a * b })
		case compiler.OpNegate:
			val := vm.pop().AsNumber()
			vm.push(value.Number(-val))
		case compiler.OpNot:
			val := vm.pop().AsBoolean()
			vm.push(value.Boolean(!val))
		case compiler.OpAdd:
			val2 := vm.pop()
			val1 := vm.pop()

			switch {
			case val1.IsNumber() && val2.IsNumber():
				vm.push(value.Number(val1.AsNumber() + val2.AsNumber()))
			default:
				vm.error("Cannot add two non-numbers")
			}
		case compiler.OpSetGlobal:
			val := vm.peek()
			i := vm.readByte()
			name := vm.chunk.Constants[i]

			vm.globals[name.String()] = val
		case compiler.OpGetGlobal:
			i := vm.readByte()
			name := vm.chunk.Constants[i]

			val := vm.globals[name.String()]

			if val == nil {
				vm.push(value.Nil{})
			} else {
				vm.push(val)
			}
		case compiler.OpJumpIfFalse:
			cond := vm.pop()

			if cond.AsBoolean() {
				vm.advance()
				vm.advance()
			} else {
				upper := byte(vm.readByte())
				lower := byte(vm.readByte())
				dist := compiler.MergeBytes(upper, lower)

				vm.ip += dist
			}
		case compiler.OpLoop:
			upper := byte(vm.readByte())
			lower := byte(vm.readByte())
			dist := compiler.MergeBytes(upper, lower)

			vm.ip -= dist
		case compiler.OpCreateTable:
			vm.push(value.NewTable())
		case compiler.OpInsertTable:
			val := vm.pop()
			table := vm.peek().AsTable()

			table.Insert(val)
		default:
			vm.error(fmt.Sprint("Do not know how to perform: ", op))
			return nil
		}
	}
}

func (vm *VM) pop() value.Value {
	vm.stackSize -= 1
	val := vm.stack[vm.stackSize]
	vm.stack[vm.stackSize] = nil

	return val
}

func (vm *VM) peek() value.Value {
	return vm.stack[vm.stackSize-1]
}

func (vm *VM) push(val value.Value) {
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
	case val1.IsNumber() && val2.IsNumber():
		vm.push(value.Number(op(val1.AsNumber(), val2.AsNumber())))
	default:
		vm.error(fmt.Sprintf("Cannot %s two non-numbers", name))
	}
}

func (vm *VM) compare(compare func(float64, float64) bool) {
	val2 := vm.pop()
	val1 := vm.pop()

	if val1.IsNumber() && val2.IsNumber() {
		vm.push(value.Boolean(compare(val1.AsNumber(), val2.AsNumber())))
	} else {
		vm.error("Unable to compare two non-numbers")
	}
}

func (vm *VM) readByte() compiler.Op {
	c := vm.current()
	vm.advance()
	return c
}

func (vm *VM) advance() {
	vm.ip += 1
}

func (vm *VM) previous() compiler.Op {
	return vm.chunk.Bytecode[vm.ip-1]
}

func (vm *VM) current() compiler.Op {
	return vm.chunk.Bytecode[vm.ip]
}

func (vm *VM) next() compiler.Op {
	return vm.chunk.Bytecode[vm.ip+1]
}

// todo: error handling
func (vm *VM) error(message string) {
	vm.err.Append(RuntimeError{message})
}

type RuntimeError struct {
	message string
}

func (re RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error ---> %s", re.message)
}
