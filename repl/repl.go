package repl

import (
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/parser"
	"BigTalk_Interpreter/vm"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Print(PROMPT)
		s := scanner.Scan()
		if !s {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.NewParser(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewCompiler()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation error:\n %s\n", err)
			continue
		}

		vMachine := vm.NewVirtualMachine(comp.ByteCode())
		err = vMachine.Run()
		if err != nil {
			fmt.Fprintf(out, "Bytecode execution error:\n %s\n", err)
			continue
		}

		top := vMachine.StackTop()
		io.WriteString(out, fmt.Sprintf("%s\n", top.Inspect()))
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! Parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", msg))
	}
}
