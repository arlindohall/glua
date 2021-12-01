package value

import "fmt"

type Value interface {
	String() string

	IsNumber() bool
	AsNumber() float64

	IsBoolean() bool
	AsBoolean() bool

	IsString() bool

	IsTable() bool
	AsTable() *Table

	IsNil() bool
}

type StringVal string

func (s StringVal) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

func (s StringVal) IsNumber() bool {
	return false
}

func (s StringVal) AsNumber() float64 {
	panic("Cannot coerce string to number")
}

func (s StringVal) IsBoolean() bool {
	return false
}

func (s StringVal) AsBoolean() bool {
	return true
}

func (s StringVal) IsString() bool {
	return true
}

func (s StringVal) IsNil() bool {
	return false
}

func (s StringVal) IsTable() bool {
	return false
}

func (s StringVal) AsTable() *Table {
	panic("Internal error: cannot cast string as table.")
}

type Number float64

func (n Number) String() string {
	return fmt.Sprint(float64(n))
}

func (n Number) IsNumber() bool {
	return true
}

func (n Number) AsNumber() float64 {
	return float64(n)
}

func (n Number) IsBoolean() bool {
	return false
}

func (n Number) AsBoolean() bool {
	return true
}

func (n Number) IsString() bool {
	return false
}

func (n Number) IsNil() bool {
	return false
}

func (n Number) IsTable() bool {
	return false
}

func (n Number) AsTable() *Table {
	panic("Internal error: cannot cast number as table.")
}

type Boolean bool

func (b Boolean) String() string {
	return fmt.Sprint(bool(b))
}

func (b Boolean) IsNumber() bool {
	return false
}

func (b Boolean) AsNumber() float64 {
	if bool(b) {
		return 1
	} else {
		return 0
	}
}

func (n Boolean) IsBoolean() bool {
	return true
}

func (b Boolean) AsBoolean() bool {
	return bool(b)
}

func (b Boolean) IsString() bool {
	return false
}

func (b Boolean) IsNil() bool {
	return false
}

func (b Boolean) IsTable() bool {
	return false
}

func (b Boolean) AsTable() *Table {
	panic("Internal error: cannot cast boolean as table.")
}

type Nil struct{}

func (n Nil) String() string {
	return "<nil>"
}

func (n Nil) IsNumber() bool {
	return false
}

func (n Nil) AsNumber() float64 {
	return 0
}

func (n Nil) IsBoolean() bool {
	return false
}

func (n Nil) AsBoolean() bool {
	return false
}

func (n Nil) IsString() bool {
	return false
}

func (n Nil) IsNil() bool {
	return true
}

func (n Nil) IsTable() bool {
	return false
}

func (n Nil) AsTable() *Table {
	panic("Internal error: cannot cast nil as table.")
}

type Table struct {
	entries map[Value]Value
}

func NewTable() *Table {
	return &Table{
		entries: make(map[Value]Value),
	}
}

func (t *Table) String() string {
	return fmt.Sprintf("Table<%p>", t)
}

func (t *Table) IsNumber() bool {
	return false
}

func (t *Table) AsNumber() float64 {
	return 0
}

func (t *Table) IsBoolean() bool {
	return false
}

func (t *Table) AsBoolean() bool {
	return false
}

func (t *Table) IsString() bool {
	return false
}

func (t *Table) IsNil() bool {
	return false
}

func (t *Table) IsTable() bool {
	return true
}

func (t *Table) AsTable() *Table {
	return t
}

func (t *Table) Set(k, v Value) {
	t.entries[k] = v
}

func (t *Table) Get(k Value) Value {
	v := t.entries[k]
	if v == nil {
		return Nil{}
	} else {
		return v
	}
}
