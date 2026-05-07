package main

import (
	"flag"
	"fmt"
	"os"
	"workflowscript/internal/lexer"
	"workflowscript/internal/parser"
)

func main() {
	dumpAST := flag.Bool("dump-ast", false, "print the AST and exit")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: wsc [--dump-ast] <file.ws>")
		os.Exit(1)
	}

	src, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	tokens, err := l.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	p := parser.New(tokens)
	prog, err := p.ParseProgram()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if *dumpAST {
		fmt.Println(prog.String())
		return
	}

	fmt.Println("Parse OK —", len(prog.Decls), "top-level declarations")
}
