package main

import (
	"fmt"
	"os"
	"tgoc/lexer"
)

func main() {
	if len(os.Args) != 2 {
		os.Exit(1)
	}

	input := os.Args[1] + "\000"

	l := lexer.New(input)
	l.Analyze()

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".globl _main\n")
	fmt.Printf("_main:\n")
	fmt.Printf("	push rbp\n")
	fmt.Printf("	mov rbp, rsp\n")
	fmt.Printf("	mov rax, %s\n", l.Tokens[0].Literal)

	for i := 1; i < len(l.Tokens)-1; i += 2 {
		if l.Tokens[i].Literal == "+" {
			fmt.Printf("	add rax, %s\n", l.Tokens[i+1].Literal)
		}
		if l.Tokens[i].Literal == "-" {
			fmt.Printf("	sub rax, %s\n", l.Tokens[i+1].Literal)
		}
	}

	fmt.Printf("	pop rbp\n")
	fmt.Printf("	ret\n")
}
