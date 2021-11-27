package value

import "fmt"

type Value interface {
	String() string

	IsNumber() bool
	AsNumber() float64

	AsBoolean() bool
}

type Number struct {
	Val float64
}

func (n Number) String() string {
	return fmt.Sprint(n.Val)
}

func (n Number) IsNumber() bool {
	return true
}

func (n Number) AsNumber() float64 {
	return n.Val
}

func (n Number) AsBoolean() bool {
	return n.Val != 0
}

type Boolean struct {
	Val bool
}

func (b Boolean) String() string {
	return fmt.Sprint(b.Val)
}

func (b Boolean) IsNumber() bool {
	return false
}

func (b Boolean) AsNumber() float64 {
	if b.Val {
		return 1
	} else {
		return 0
	}
}

func (b Boolean) AsBoolean() bool {
	return b.Val
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

func (n Nil) AsBoolean() bool {
	return false
}
