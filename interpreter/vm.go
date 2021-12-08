package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/glerror"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

type CallFrame struct {
	ip      int
	stack   int
	closure *value.Closure
	context *CallFrame
}

type VM struct {
	frame     *CallFrame
	stack     []value.Value
	stackSize int
	globals   map[string]value.Value
	err       glerror.GluaErrorChain
}

func NewVm() VM {
	return VM{
		frame:     nil,
		stack:     nil,
		stackSize: 0,
		globals:   make(map[string]value.Value),
		err:       glerror.GluaErrorChain{},
	}
}

func (vm *VM) Interpret(chunk value.Chunk) (value.Value, glerror.GluaErrorChain) {
	closure := value.Closure{
		Name:  "",
		Chunk: chunk,
	}

	vm.push(&closure)
	vm.push(value.Number(0))
	vm.call()

	val := vm.run()

	return val, vm.err
}

func (vm *VM) run() value.Value {
	for {
		op := vm.readByte()

		if TraceExecution {
			DebugTrace(vm)
		}

		var ok bool = true

		switch op {
		case compiler.OpAssert:
			val := vm.pop()
			if !val.AsBoolean() {
				// exit or break to top level?
				os.Exit(5)
			}
		case compiler.OpPop:
			vm.pop()
		case compiler.OpConstant:
			c := byte(vm.readByte())
			val := vm.getConstant(c)

			vm.push(val)
		case compiler.OpNil:
			vm.push(value.Nil{})
		case compiler.OpZero:
			vm.push(value.Number(0))
		case compiler.OpLessThan:
			ok = vm.compare(func(v1, v2 float64) bool { return v1 < v2 })
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
			ok = vm.arithmetic("subtract", func(a, b float64) float64 { return a - b })
		case compiler.OpDivide:
			ok = vm.arithmetic("divide", func(a, b float64) float64 { return a / b })
		case compiler.OpMult:
			ok = vm.arithmetic("multiply", func(a, b float64) float64 { return a * b })
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
				return vm.error("Cannot add two non-numbers")
			}
		case compiler.OpSetGlobal:
			val := vm.peek()
			i := vm.readByte()
			name := vm.frame.closure.Chunk.Constants[i]

			vm.globals[name.String()] = val
		case compiler.OpGetGlobal:
			i := vm.readByte()
			name := vm.frame.closure.Chunk.Constants[i].String()

			val := vm.globals[name]

			if val == nil {
				vm.push(value.Nil{})
			} else {
				vm.push(val)
			}
		case compiler.OpGetLocal:
			slot := vm.readByte()
			val := vm.getLocal(byte(slot))

			vm.push(val)
		case compiler.OpSetLocal:
			slot := vm.readByte()
			val := vm.peek()

			vm.setLocal(byte(slot), val)
		case compiler.OpJumpIfFalse:
			cond := vm.pop()

			if cond.AsBoolean() {
				vm.advance()
				vm.advance()
			} else {
				upper := byte(vm.readByte())
				lower := byte(vm.readByte())
				dist := compiler.MergeBytes(upper, lower)

				vm.frame.ip += dist
			}
		case compiler.OpLoop:
			upper := byte(vm.readByte())
			lower := byte(vm.readByte())
			dist := compiler.MergeBytes(upper, lower)

			vm.frame.ip -= dist
		case compiler.OpCreateTable:
			vm.push(value.NewTable())
		case compiler.OpInsertTable:
			val := vm.pop()
			table := vm.peek().AsTable()

			table.Insert(val)
		case compiler.OpSetTable:
			val := vm.pop()
			key := vm.pop()
			table := vm.pop().AsTable()

			ok := table.Set(key, val)

			if !ok {
				return vm.error("Cannot set key <nil> in table.")
			}

			vm.push(val)
		case compiler.OpInitTable:
			// Exact same as set table, but leaves table on stack instead of value
			val := vm.pop()
			key := vm.pop()
			table := vm.peek().AsTable()

			ok := table.Set(key, val)

			if !ok {
				return vm.error("Cannot set key <nil> in table.")
			}
		case compiler.OpGetTable:
			attribute := vm.pop()
			table := vm.pop()

			if !table.IsTable() {
				return vm.error("Cannot assign to non-table")
			}

			val := table.AsTable().Get(attribute)
			vm.push(val)
		case compiler.OpCall:
			vm.call()
		case compiler.OpReturn:
			// todo: multiple return
			if vm.frame.context == nil {
				return vm.pop()
			} else {
				vm.returnFrom()
			}
		default:
			return vm.error(fmt.Sprint("Do not know how to perform: ", op))
		}

		if !ok {
			return value.Nil{}
		}
	}
}

func (vm *VM) call() {
	arity := int(vm.pop().AsNumber())
	stackBottom := vm.stackSize - arity - 1
	closure := vm.stack[stackBottom].AsFunction()
	enclosing := vm.frame
	frame := CallFrame{
		ip:      0,
		stack:   stackBottom,
		context: enclosing,
		closure: closure,
	}
	vm.frame = &frame

	vm.traceFunction()
}

func (vm *VM) returnFrom() {
	val := vm.pop()

	stack := vm.frame.stack
	context := vm.frame.context

	vm.frame = context
	vm.stackSize = stack

	for i := stack; i < vm.stackSize; i++ {
		vm.stack[i] = nil
	}

	vm.push(val)

	vm.traceFunction()
}

func (vm *VM) traceFunction() {
	// todo: call function
	if TraceExecution && vm.frame.closure.Name == "" {
		fmt.Fprintln(os.Stderr, "========== <script> ==========")
	} else if TraceExecution {
		fmt.Fprintf(os.Stderr, "========== %s ==========\n", vm.frame.closure.Name)
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

func (vm *VM) getConstant(slot byte) value.Value {
	return vm.frame.closure.Chunk.Constants[slot]
}

func (vm *VM) getLocal(slot byte) value.Value {
	return vm.stack[slot]
}

func (vm *VM) setLocal(slot byte, val value.Value) {
	vm.stack[slot] = val
}

func (vm *VM) arithmetic(name string, op func(float64, float64) float64) bool {
	val2 := vm.pop()
	val1 := vm.pop()

	switch {
	case val1.IsNumber() && val2.IsNumber():
		vm.push(value.Number(op(val1.AsNumber(), val2.AsNumber())))
		return true
	default:
		vm.error(fmt.Sprintf("Cannot %s two non-numbers", name))
		return false
	}
}

func (vm *VM) compare(compare func(float64, float64) bool) bool {
	val2 := vm.pop()
	val1 := vm.pop()

	if val1.IsNumber() && val2.IsNumber() {
		vm.push(value.Boolean(compare(val1.AsNumber(), val2.AsNumber())))
		return true
	} else {
		vm.error("Unable to compare two non-numbers")
		return false
	}
}

func (vm *VM) readByte() byte {
	c := vm.current()
	vm.advance()
	return c
}

func (vm *VM) advance() {
	vm.frame.ip += 1
}

func (vm *VM) previous() byte {
	return vm.frame.closure.Chunk.Bytecode[vm.frame.ip-1]
}

func (vm *VM) current() byte {
	return vm.frame.closure.Chunk.Bytecode[vm.frame.ip]
}

func (vm *VM) next() byte {
	return vm.frame.closure.Chunk.Bytecode[vm.frame.ip+1]
}

func (vm *VM) error(message string) value.Value {
	vm.err.Append(RuntimeError{
		message: message,
		line:    vm.frame.closure.Chunk.Lines[vm.frame.ip],
	})
	return value.Nil{}
}

type RuntimeError struct {
	message string
	line    int
}

func (re RuntimeError) Error() string {
	return fmt.Sprintf("Runtime error [line=%d] ---> %s", re.line, re.message)
}

func (vm *VM) GetErrors() error {
	return vm.err
}

func (vm *VM) ClearErrors() {
	vm.err = glerror.GluaErrorChain{}
}
