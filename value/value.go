package value

import "fmt"

type Value interface {
	String() string

	IsNumber() bool
	AsNumber() float64

	IsBoolean() bool
	AsBoolean() bool

	IsString() bool

	IsNil() bool
}

type StringVal string

func (s StringVal) String() string {
	return string(s)
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
	return float64(n) != 0
}

func (n Number) IsString() bool {
	return false
}

func (n Number) IsNil() bool {
	return false
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
