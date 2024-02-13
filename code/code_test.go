package code

import "testing"

func TestMakeInstruction(t *testing.T) {
	testCases := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{
			OpConstant,
			[]int{65534},
			[]byte{byte(OpConstant), 255, 254},
		},
	}

	for _, tc := range testCases {
		instruction := MakeInstruction(tc.op, tc.operands...)

		if len(instruction) != len(tc.expected) {
			t.Errorf("len(instruction) = %d, want %d", len(instruction), len(tc.expected))
		}

		for i, b := range tc.expected {
			if instruction[i] != tc.expected[i] {
				t.Errorf("wrong byte at position %d, got = %d, want = %d", i, instruction[i], b)
			}
		}
	}
}
