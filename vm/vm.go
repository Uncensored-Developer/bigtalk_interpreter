package vm

import (
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/object"
	"fmt"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536 // maximum number of global binding the VM can support
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VirtualMachine struct {
	constants    []object.IObject
	instructions code.Instructions
	stack        []object.IObject
	sp           int // Points to the next value

	globals []object.IObject
}

func NewVirtualMachine(bytecode *compiler.ByteCode) *VirtualMachine {
	return &VirtualMachine{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.IObject, StackSize),
		sp:           0,
		globals:      make([]object.IObject, GlobalsSize),
	}
}

func NewVirtualMachineWithGlobalStore(bytecode *compiler.ByteCode, s []object.IObject) *VirtualMachine {
	vm := NewVirtualMachine(bytecode)
	vm.globals = s
	return vm
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
		case code.OpBang:
			err := v.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := v.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			// Decode the operand on the right after the Opcode
			pos := int(code.ReadUint16(v.instructions[i+1:]))
			// Set the instruction pointer to the target of our jump
			// accounting for the default increment of i in the for-loop
			i = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(v.instructions[i+1:]))
			i += 2

			condition := v.pop()
			if !isTruthy(condition) {
				i = pos - 1
			}
		case code.OpNull:
			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(v.instructions[i+1:])
			i += 2
			v.globals[globalIdx] = v.pop()
		case code.OpGetGlobal:
			globalIdx := code.ReadUint16(v.instructions[i+1:])
			i += 2
			err := v.push(v.globals[globalIdx])
			if err != nil {
				return err
			}
		case code.OpArray:
			arrayLength := int(code.ReadUint16(v.instructions[i+1:]))
			i += 2

			array := v.buildArray(v.sp-arrayLength, v.sp)
			v.sp = v.sp - arrayLength

			err := v.push(array)
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

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return v.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return v.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
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

func (v *VirtualMachine) executeBinaryStringOperation(op code.Opcode, left, right object.IObject) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return v.push(&object.String{Value: leftValue + rightValue})
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

func (v *VirtualMachine) executeBangOperator() error {
	operand := v.pop()

	switch operand {
	case True:
		return v.push(False)
	case False:
		return v.push(True)
	case Null:
		return v.push(True)
	default:
		return v.push(False)
	}
}

func (v *VirtualMachine) executeMinusOperator() error {
	operand := v.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for -: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return v.push(&object.Integer{Value: -value})
}

func (v *VirtualMachine) buildArray(startIndex, endIndex int) object.IObject {
	items := make([]object.IObject, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		items[i-startIndex] = v.stack[i]
	}
	return &object.Array{Items: items}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func isTruthy(obj object.IObject) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}
