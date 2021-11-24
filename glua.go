package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Running REPL...")
		return
	}

	fileName := os.Args[1]

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("Error opening file", fileName, err)
		return
	}

	reader := bufio.NewReader(file)

	scanner := Scanner(reader)
	tokens, err := scanner.ScanTokens()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(tokens)
}
