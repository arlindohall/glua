package main

import "fmt"

type value interface {
	String() string

	isNumber() bool
	asNumber() float64
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
