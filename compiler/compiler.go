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

type EmittedInstructions struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.IObject
	lastInstruction     EmittedInstructions
	previousInstruction EmittedInstructions
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.IObject{},
		lastInstruction:     EmittedInstructions{},
		previousInstruction: EmittedInstructions{},
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

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		afterConsequencePosition := len(c.instructions)
		c.changeOperand(jumpNotTruthyPosition, afterConsequencePosition)

		if node.Alternative == nil {
			afterConsequencePosition := len(c.instructions)
			c.changeOperand(jumpNotTruthyPosition, afterConsequencePosition)
		} else {
			jumpPosition := c.emit(code.OpJump, 999)

			afterConsequencePosition := len(c.instructions)
			c.changeOperand(jumpNotTruthyPosition, afterConsequencePosition)

			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

			afterAlternativePosition := len(c.instructions)
			c.changeOperand(jumpPosition, afterAlternativePosition)
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
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
	c.setLastInstructions(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstructions(op code.Opcode, position int) {
	previous := c.lastInstruction
	last := EmittedInstructions{Opcode: op, Position: position}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

// replaceInstruction replaces the instruction at the specified position in the Compiler's instruction list with the new instruction.
// It iterates over each byte in the new instruction and updates the corresponding byte in the instruction list.
func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[position+i] = newInstruction[i]
	}
}

// changeOperand takes an opcode position and an operand value and replaces the existing instruction at that position with a new instruction generated using the provided opcode and operand
func (c *Compiler) changeOperand(opPosition int, operand int) {
	op := code.Opcode(c.instructions[opPosition])
	newInstruction := code.MakeInstruction(op, operand)
	c.replaceInstruction(opPosition, newInstruction)
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}
