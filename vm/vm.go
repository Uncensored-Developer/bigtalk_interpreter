package vm

import (
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/compiler"
	"BigTalk_Interpreter/object"
	"fmt"
)

const StackSize = 2048

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

func (v *VirtualMachine) StackTop() object.IObject {
	if v.sp == 0 {
		return nil
	}
	return v.stack[v.sp-1]
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
