package main

import "fmt"

type value interface {
	String() string
}

type number struct {
	val float64
}

func (n *number) String() string {
	return fmt.Sprint(n.val)
}
