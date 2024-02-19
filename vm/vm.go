package vm

import (
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/object"
	"fmt"
)

const StackSize = 2048

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
)

type VirtualMachine struct {
	constants    []object.IObject
	instructions code.Instructions
	stack        []object.IObject
	sp           int // Points to the next value
}

func NewVirtualMachine(bytecode *compiler.ByteCode) *VirtualMachine {
	return &VirtualMachine{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.IObject, StackSize),
		sp:           0,
	}
}

func (v *VirtualMachine) LastPoppedStackElement() object.IObject {
	return v.stack[v.sp]
}

func (v *VirtualMachine) Run() error {
	for i := 0; i < len(v.instructions); i++ {
		op := code.Opcode(v.instructions[i])
		switch op {
		case code.OpConstant:
			index := code.ReadUint16(v.instructions[i+1:])
			i += 2
			err := v.push(v.constants[index])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := v.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			v.pop()
		case code.OpTrue:
			err := v.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := v.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := v.executeComparison(op)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VirtualMachine) push(obj object.IObject) error {
	if v.sp >= StackSize {
		return fmt.Errorf("stack overflow error")
	}
	v.stack[v.sp] = obj
	v.sp++
	return nil
}

func (v *VirtualMachine) pop() object.IObject {
	obj := v.stack[v.sp-1]
	v.sp--
	return obj
}

func (v *VirtualMachine) executeBinaryOperation(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return v.executeBinaryIntegerOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (v *VirtualMachine) executeBinaryIntegerOperation(op code.Opcode, left, right object.IObject) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return v.push(&object.Integer{Value: result})
}

func (v *VirtualMachine) executeComparison(op code.Opcode) error {
	right := v.pop()
	left := v.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return v.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (v *VirtualMachine) executeIntegerComparison(op code.Opcode, left, right object.IObject) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return v.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return v.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return v.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}
