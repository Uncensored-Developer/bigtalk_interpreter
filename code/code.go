package code

import (
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type Opcode byte

// Opcodes definitions
const (
	OpConstant Opcode = iota
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
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}
	return instruction
}
