package compiler

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/object"
	"fmt"
	"sort"
)

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.IObject
}

type EmittedInstructions struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	constants []object.IObject

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

func NewCompiler() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstructions{},
		previousInstruction: EmittedInstructions{},
	}
	return &Compiler{
		constants:   []object.IObject{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}

}

func NewCompilerWithState(s *SymbolTable, constants []object.IObject) *Compiler {
	compiler := NewCompiler()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) Compile(node ast.INode) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.RightValue)
			if err != nil {
				return err
			}
			err = c.Compile(node.LeftValue)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(node.LeftValue)
		if err != nil {
			return err
		}

		err = c.Compile(node.RightValue)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("invalid operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPosition := c.emit(code.OpJumpNotTruthy, 999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPosition := c.emit(code.OpJump, 999)
		afterConsequencePosition := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPosition, afterConsequencePosition)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}
		afterAlternativePosition := len(c.currentInstructions())
		c.changeOperand(jumpPosition, afterAlternativePosition)
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		if symbol.Scope == GlobalScope {
			c.emit(code.OpGetGlobal, symbol.Index)
		} else {
			c.emit(code.OpGetLocal, symbol.Index)
		}
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, item := range node.Items {
			err := c.Compile(item)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Items))
	case *ast.MapLiteral:
		var keys []ast.IExpression
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpMap, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		localsCount := c.symbolTable.numDefinitions
		instructions := c.leaveScope()
		compiledFn := &object.CompiledFunction{
			Instructions:    instructions,
			LocalsCount:     localsCount,
			ParametersCount: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFn))
	case *ast.ReturnStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Func)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))
	}
	return nil
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.IObject) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit creates a bytecode instruction with the given opcode and operands,
// appends it to the list of instructions, and returns its position.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.MakeInstruction(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstructions(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstructions(op code.Opcode, position int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstructions{Opcode: op, Position: position}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	prev := c.scopes[c.scopeIndex].previousInstruction

	oldIns := c.currentInstructions()
	newIns := oldIns[:last.Position]

	c.scopes[c.scopeIndex].instructions = newIns
	c.scopes[c.scopeIndex].lastInstruction = prev
}

// replaceInstruction replaces the instruction at the specified position in the Compiler's instruction list with the new instruction.
// It iterates over each byte in the new instruction and updates the corresponding byte in the instruction list.
func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[position+i] = newInstruction[i]
	}
}

// changeOperand takes an opcode position and an operand value and replaces the existing instruction at that position with a new instruction generated using the provided opcode and operand
func (c *Compiler) changeOperand(opPosition int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPosition])
	newInstruction := code.MakeInstruction(op, operand)
	c.replaceInstruction(opPosition, newInstruction)
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstructions{},
		previousInstruction: EmittedInstructions{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewWrappedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	ins := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
	return ins
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.MakeInstruction(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstructions
	previousInstruction EmittedInstructions
}
