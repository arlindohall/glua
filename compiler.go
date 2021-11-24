package main

const (
	OpNil = iota
	OpReturn
)

type compiler struct {
	text     []Token
	bytecode []op
}

type Function struct {
	bytecode []op
	name     string
}

func Compile(text []Token) Function {
	compiler := compiler{
		text,
		nil,
	}

	return compiler.end()
}

func (compiler *compiler) emit(b1, b2 byte) {
	compiler.bytecode = append(compiler.bytecode, op(b1))
	compiler.bytecode = append(compiler.bytecode, op(b2))
}

func (compiler *compiler) emitReturn() {
	compiler.emit(OpNil, OpReturn)
}

func (compiler *compiler) end() Function {
	compiler.emitReturn()

	return Function{
		compiler.bytecode,
		"",
	}
}
