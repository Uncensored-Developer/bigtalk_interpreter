package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *OpcodeDefinition, operands []int) string {
	operandsCount := len(def.OperandWidths)

	if len(operands) != operandsCount {
		return fmt.Sprintf("ERROR: operand count %d is not the defined %d", len(operands), operandsCount)
	}

	switch operandsCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

// Opcodes definitions
const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpBang
	OpMinus
	OpJump
	OpJumpNotTruthy
	OpNull
	OpSetGlobal
	OpGetGlobal
	OpArray
	OpMap
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpSetLocal
	OpGetLocal
	OpGetBuiltin
	OpClosure
)

type OpcodeDefinition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*OpcodeDefinition{
	OpConstant: {
		Name:          "OpConstant",
		OperandWidths: []int{2}, // 2 bytes wide
	},
	OpAdd: {
		Name:          "OpAdd",
		OperandWidths: []int{},
	},
	OpPop: {
		Name:          "OpPop",
		OperandWidths: []int{},
	},
	OpSub: {
		Name:          "OpSub",
		OperandWidths: []int{},
	},
	OpMul: {
		Name:          "OpMul",
		OperandWidths: []int{},
	},
	OpDiv: {
		Name:          "OpDiv",
		OperandWidths: []int{},
	},
	OpTrue: {
		Name:          "OpTrue",
		OperandWidths: []int{},
	},
	OpFalse: {
		Name:          "OpFalse",
		OperandWidths: []int{},
	},
	OpEqual: {
		Name:          "OpEqual",
		OperandWidths: []int{},
	},
	OpNotEqual: {
		Name:          "OpNotEqual",
		OperandWidths: []int{},
	},
	OpGreaterThan: {
		Name:          "OpGreaterThan",
		OperandWidths: []int{},
	},
	OpBang: {
		Name:          "OpBang",
		OperandWidths: []int{},
	},
	OpMinus: {
		Name:          "OpMinus",
		OperandWidths: []int{},
	},
	OpJump: {
		Name:          "OpJump",
		OperandWidths: []int{2},
	},
	OpJumpNotTruthy: {
		Name:          "OpJumpNotTruthy",
		OperandWidths: []int{2},
	},
	OpNull: {
		Name:          "OpNull",
		OperandWidths: []int{},
	},
	OpSetGlobal: {
		Name:          "OpSetGlobal",
		OperandWidths: []int{2},
	},
	OpGetGlobal: {
		Name:          "OpGetGlobal",
		OperandWidths: []int{2},
	},
	OpArray: {
		Name:          "OpArray",
		OperandWidths: []int{2}, // this means an array literal can contain a maximum of 65535 items. (2 bytes)
	},
	OpMap: {
		Name:          "OpMap",
		OperandWidths: []int{2},
	},
	OpIndex: {
		Name:          "OpIndex",
		OperandWidths: []int{},
	},
	OpCall: {
		Name:          "OpCall",
		OperandWidths: []int{1},
	},
	OpReturnValue: {
		Name:          "OpReturnValue",
		OperandWidths: []int{},
	},
	OpReturn: {
		Name:          "OpReturn",
		OperandWidths: []int{},
	},
	OpSetLocal: {
		Name:          "OpSetLocal",
		OperandWidths: []int{1},
	},
	OpGetLocal: {
		Name:          "OpGetLocal",
		OperandWidths: []int{1},
	},
	OpGetBuiltin: {
		Name:          "OpGetBuiltin",
		OperandWidths: []int{1},
	},
	OpClosure: {
		Name:          "OpClosure",
		OperandWidths: []int{2, 1},
	},
}

func Lookup(op byte) (*OpcodeDefinition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// MakeInstruction Creates a single bytecode instruction that's made up of an Opcode
// and an optional number of arguments.
func MakeInstruction(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, operand := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		case 1:
			instruction[offset] = byte(operand)
		}
		offset += width
	}
	return instruction
}

// ReadOperands reads the operands of an instruction based on the given OpcodeDefinition and Instructions.
// It returns the operands as a slice of integers and the number of bytes read.
func ReadOperands(def *OpcodeDefinition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
