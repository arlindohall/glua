package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Running REPL...")
		return
	}

	fmt.Println("Running program", os.Args[1])
}
