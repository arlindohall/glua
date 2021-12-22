package interpreter

import (
	"arlindohall/glua/compiler"
	"arlindohall/glua/constants"
	"arlindohall/glua/glerror"
	"arlindohall/glua/value"
	"fmt"
	"os"
)

type CallFrame struct {
	ip           int
	stack        int
	closure      *value.Closure
	context      *CallFrame
	isAssignment bool
}

type VM struct {
	frame        *CallFrame
	assignBase   []int
	assignTarget []int
	localTarget  []int
	stackSize    int
	stack        []value.Value
	openUpvalues []*value.Upvalue
	globals      map[string]value.Value
	err          glerror.GluaErrorChain
}

func NewVm() VM {
	vm := VM{
		frame:     nil,
		stack:     nil,
		stackSize: 0,
		globals:   make(map[string]value.Value),
		err:       glerror.GluaErrorChain{},
	}

	vm.addBuiltins()

	return vm
}

func (vm *VM) addBuiltins() {
	vm.globals["time"] = value.NewBuiltin("time", value.Time)
}

func (vm *VM) Interpret(function compiler.Function) (value.Value, glerror.GluaErrorChain) {
	closure := value.NewClosure(function.Chunk, function.Name)

	vm.push(closure)
	vm.call(0, false)

	val := vm.run()

	return val, vm.err
}

func (vm *VM) run() value.Value {
	for {
		op := vm.readByte()

		if constants.TraceExecution {
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
		case compiler.OpLess:
			ok = vm.compare(func(v1, v2 float64) bool { return v1 < v2 })
		case compiler.OpGreater:
			ok = vm.compare(func(v1, v2 float64) bool { return v1 > v2 })
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
		case compiler.OpAssignStart:
			vm.addAssignment(vm.stackSize)
		case compiler.OpAssignCleanup:
			// for security, clear stack
			vm.clearStack(vm.popAssignment())
		case compiler.OpLocalAllocate:
			// todo: handle if there are fewer values than we need
			vm.addLocalAssignment(vm.stackSize + int(vm.readByte()))
		case compiler.OpLocalCleanup:
			// stack=[x, y] stackSize=2
			// z, w = 1, 2, 3
			// stack=[x, y, 1, 2, 3] vm.localTarget=(2+2)=4
			// desired stack=[x, y, 1, 2]
			vm.clearStack(vm.popLocalAssignment())
		case compiler.OpSetGlobal:
			val := vm.getAssign()
			i := vm.readByte()
			name := vm.frame.closure.Chunk.Constants[i]

			vm.globals[name.RawString()] = val
		case compiler.OpGetGlobal:
			i := vm.readByte()
			name := vm.frame.closure.Chunk.Constants[i].RawString()

			val := vm.globals[name]

			if val == nil {
				vm.push(value.Nil{})
			} else {
				vm.push(val)
			}
		case compiler.OpSetLocal:
			slot := vm.readByte()
			val := vm.getAssign()

			vm.setLocal(byte(slot), val)
		case compiler.OpGetLocal:
			slot := vm.readByte()
			val := vm.getLocal(byte(slot))

			vm.push(val)
		case compiler.OpCreateUpvalue:
			index := vm.readByte()
			isLocal := vm.readByte() == 1
			closure := vm.peek().AsClosure()

			vm.createUpvalue(index, isLocal, closure)
		case compiler.OpSetUpvalue:
			index := vm.readByte()
			val := vm.getAssign()

			vm.setUpvalue(index, val)
		case compiler.OpGetUpvalue:
			// todo: trace isLocal too
			index := vm.readByte()
			val := vm.getUpvalue(index)

			vm.push(val)
		case compiler.OpCloseUpvalues:
			index := vm.readByte()
			vm.closeUpvalues(vm.frame.stack + int(index))
			vm.clearStack(vm.frame.stack + int(index))
		case compiler.OpClosure:
			// Copy closure
			closure := vm.pop().AsClosure()

			vm.push(&value.Closure{
				Chunk:    closure.Chunk,
				Name:     closure.Name,
				Upvalues: nil,
			})
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
			val := vm.getAssign()
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
			arity := int(vm.readByte())
			isAssignment := vm.readByte() == 1
			vm.call(arity, isAssignment)
		case compiler.OpReturn:
			arity := int(vm.readByte())
			vm.returnFrom(arity)

			if vm.frame == nil {
				return vm.pop()
			}
		default:
			return vm.error(fmt.Sprint("Do not know how to perform: ", compiler.ByteName(op)))
		}

		if !ok {
			return value.Nil{}
		}
	}
}

func (vm *VM) addAssignment(capacity int) {
	vm.assignBase = append(vm.assignBase, capacity)
	vm.assignTarget = append(vm.assignTarget, capacity)
}

func (vm *VM) incAssignTarget() int {
	index := len(vm.assignTarget) - 1
	result := vm.assignTarget[index]
	vm.assignTarget[index] += 1

	return result
}

func (vm *VM) popAssignment() int {
	assign := vm.assignBase[len(vm.assignBase)-1]
	vm.assignBase = vm.assignBase[1:]
	vm.assignTarget = vm.assignTarget[1:]

	return assign
}

func (vm *VM) addLocalAssignment(capacity int) {
	vm.localTarget = append(vm.localTarget, capacity)
}

func (vm *VM) popLocalAssignment() int {
	index := len(vm.localTarget) - 1
	assign := vm.localTarget[index]
	vm.localTarget = vm.localTarget[:index]

	return assign
}

func (vm *VM) call(arity int, isAssignment bool) {
	// stack=[x, y, func, a, b, c]; stackSize=6; arity=3 -> stackBottom=2
	stackBottom := vm.stackSize - arity - 1

	if vm.stack[stackBottom].IsClosure() {
		closure := vm.stack[stackBottom].AsClosure()
		enclosing := vm.frame
		frame := CallFrame{
			ip:           0,
			stack:        stackBottom,
			context:      enclosing,
			closure:      closure,
			isAssignment: isAssignment,
		}
		vm.frame = &frame

		vm.traceFunction()
	} else if vm.stack[stackBottom].IsBuiltin() {
		arguments := vm.stack[stackBottom+1 : vm.stackSize]
		for range arguments {
			vm.pop()
		}

		builtin := vm.pop().AsBuiltin()
		vm.push(builtin.Function(arguments))
	}
}

// todo: leaving nil values on the stack somehow?
func (vm *VM) returnFrom(arity int) {
	values := make([]value.Value, arity)

	for i := 1; i <= arity; i++ {
		values[arity-i] = vm.pop()
	}

	// stack=[x, y, func, a, b, c, r1, r2, r3]; frame.stack=2; arity=3
	vm.closeUpvalues(vm.frame.stack)

	// stack=[x, y, func, a, b, c, r1, r2, r3]; frame.stack=2; arity=3
	stack := vm.frame.stack
	context := vm.frame.context
	isAssignment := vm.frame.isAssignment

	// remove all stack entries from stack..stackSize, not inclusive of stackSize
	// this also sets the stack top to the stack top before the call, dropping
	// the closure as well as parameters and locals
	//
	// then set the call frame to the parent call frame
	// stack=[x, y]; frame.stack=0
	vm.clearStack(stack)
	vm.frame = context

	if isAssignment && len(values) > 0 {
		for _, value := range values {
			vm.push(value)
		}
	} else if isAssignment && len(values) == 0 {
		vm.push(value.Nil{})
	} else {
		vm.push(values[0])
	}

	if vm.frame != nil {
		vm.traceFunction()
	}
}

func (vm *VM) clearStack(stack int) {
	for i := stack; i < vm.stackSize; i++ {
		vm.stack[i] = nil
	}
	vm.stackSize = stack
}

func (vm *VM) traceFunction() {
	// todo: call function
	if constants.TraceExecution && vm.frame.closure.Name == "" {
		fmt.Fprintln(os.Stderr, "========== <script> ==========")
	} else if constants.TraceExecution {
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

func (vm *VM) getAssign() value.Value {
	index := vm.incAssignTarget()
	if index >= vm.stackSize {
		return value.Nil{}
	} else {
		return vm.stack[index]
	}
}

func (vm *VM) getLocal(slot byte) value.Value {
	return vm.stack[vm.frame.stack+int(slot)]
}

func (vm *VM) setLocal(slot byte, val value.Value) {
	vm.stack[vm.frame.stack+int(slot)] = val
}

func (vm *VM) createUpvalue(index byte, isLocal bool, closure *value.Closure) {
	var upvalue *value.Upvalue

	if isLocal {
		upvalue = &value.Upvalue{
			Value:   nil,
			Pointer: &vm.stack[vm.frame.stack+int(index)],
			IsLocal: true,
			Index:   int(index),
		}

	} else {
		root := vm.frame.closure.Upvalues[index]
		upvalue = &value.Upvalue{
			Value:   nil,
			Pointer: root.Pointer,
			IsLocal: false,
			Index:   root.Index,
		}
	}

	vm.addOpenUpvalue(upvalue)

	closure.Upvalues = append(closure.Upvalues, upvalue)
}

func (vm *VM) addOpenUpvalue(upvalue *value.Upvalue) {
	vm.openUpvalues = append(vm.openUpvalues, upvalue)
}

func (vm *VM) getUpvalue( /*frame *CallFrame,*/ index byte) value.Value {
	return *vm.frame.closure.Upvalues[index].Pointer

}

func (vm *VM) setUpvalue(index byte, val value.Value) {
	*vm.frame.closure.Upvalues[index].Pointer = val
}

func (vm *VM) closeUpvalues(index int) {
	dropping, keeping := vm.partitionUpvalues(index)
	vm.openUpvalues = keeping

	for _, upvalue := range dropping {
		upvalue.Close()
	}
}

func (vm *VM) partitionUpvalues(index int) (keeping []*value.Upvalue, dropping []*value.Upvalue) {
	for _, upvalue := range vm.openUpvalues {
		if upvalue.Index > index {
			dropping = append(dropping, upvalue)
		} else {
			keeping = append(keeping, upvalue)
		}
	}

	return
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
