package main

import "fmt"

type Value interface {
	String() string

	isNumber() bool
	asNumber() float64

	asBoolean() bool
}

type number struct {
	val float64
}

func (n number) String() string {
	return fmt.Sprint(n.val)
}

func (n number) isNumber() bool {
	return true
}

func (n number) asNumber() float64 {
	return n.val
}

func (n number) asBoolean() bool {
	return n.val != 0
}

type boolean struct {
	val bool
}

func (b boolean) String() string {
	return fmt.Sprint(b.val)
}

func (b boolean) isNumber() bool {
	return false
}

func (b boolean) asNumber() float64 {
	if b.val {
		return 1
	} else {
		return 0
	}
}

func (b boolean) asBoolean() bool {
	return b.val
}

type nilVal struct{}

func (n nilVal) String() string {
	return "<nil>"
}

func (n nilVal) isNumber() bool {
	return false
}

func (n nilVal) asNumber() float64 {
	return 0
}

func (n nilVal) asBoolean() bool {
	return false
}
