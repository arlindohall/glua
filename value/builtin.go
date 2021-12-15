package value

import (
	"time"
)

func NewBuiltin(name string, f BuiltinFunc) *Builtin {
	return &Builtin{
		Name:     name,
		Function: f,
	}
}

func Time(args []Value) Value {
	return Number(time.Now().UnixNano())
}
