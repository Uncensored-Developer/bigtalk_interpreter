package compiler

import (
	"BigTalk_Interpreter/ast"
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/object"
	"fmt"
)

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.IObject
}

type Compiler struct {
	instructions code.Instructions
	constants    []object.IObject
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.IObject{},
	}
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
	case *ast.InfixExpression:
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
			c.emit(code.OppAdd)
		default:
			return fmt.Errorf("invalid operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	}
	return nil
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.instructions,
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
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
