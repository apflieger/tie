package main

import (
	"fmt"
	"os"
)

func main() {
	//repo, _ := git.OpenRepository(".")
	verb := os.Args[1]
	args := os.Args[2:]
	fmt.Printf("verb: %v, args:%v\n", verb, args)
}
