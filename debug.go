package main

import "fmt"

func (tt TokenType) String() string {
	switch tt {
	case TokenNumber:
		return "TokenNumber"
	case TokenEof:
		return "TokenEof"
	case TokenPlus:
		return "TokenPlus"
	case TokenSemicolon:
		return "TokenSemicolon"
	default:
		panic("Unrecognized TokenType")
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v/\"%v\"", t._type, t.text)
}

func (op op) String() string {
	switch op {
	case OpNil:
		return "OpNil"
	case OpConstant:
		return "OpConstant"
	case OpReturn:
		return "OpReturn"
	case OpAdd:
		return "OpAdd"
	case OpMult:
		return "OpMult"
	default:
		panic(fmt.Sprint("Unrecognized Op: ", byte(op)))
	}
}

func debugTrace(vm *VM) {
	fmt.Println("Running command", vm.current())
}

func debugPrint(chunk chunk) {
	i := 0
	var print func(int, []op) int
	for i < len(chunk.bytecode) {
		switch chunk.bytecode[i] {
		case OpAdd:
			print = printInstruction
		case OpConstant:
			print = printConstant
		case OpNil:
			print = printInstruction
		case OpReturn:
			print = printInstruction
		default:
			panic(fmt.Sprint("Unknown op for debug print: ", chunk.bytecode[i]))
		}
		i = print(i, chunk.bytecode)
	}
}

func printInstruction(i int, bytecode []op) int {
	fmt.Printf("%04d | %-12v\n", i, bytecode[i])
	return i + 1
}

func printConstant(i int, bytecode []op) int {
	fmt.Printf("%04d | %-12v %-4d\n", i, bytecode[i], bytecode[i+1])
	return i + 2
}

func debugTokens(tokens []Token) {
	for _, token := range tokens {
		if token._type == TokenSemicolon {
			fmt.Println(";")
			continue
		}
		fmt.Print(token._type, " ")
	}
}
