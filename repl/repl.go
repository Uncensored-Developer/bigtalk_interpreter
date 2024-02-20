package repl

import (
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/lexer"
	"BigTalk_Interpreter/object"
	"BigTalk_Interpreter/parser"
	"BigTalk_Interpreter/vm"
	"bufio"
	"fmt"
	"io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	var constants []object.IObject
	globals := make([]object.IObject, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

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

		comp := compiler.NewCompilerWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation error:\n %s\n", err)
			continue
		}

		vMachine := vm.NewVirtualMachineWithGlobalStore(comp.ByteCode(), globals)
		err = vMachine.Run()
		if err != nil {
			fmt.Fprintf(out, "Bytecode execution error:\n %s\n", err)
			continue
		}

		top := vMachine.LastPoppedStackElement()
		io.WriteString(out, fmt.Sprintf("%s\n", top.Inspect()))
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! Parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, fmt.Sprintf("\t%s\n", msg))
	}
}
