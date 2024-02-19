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
		{
			OpAdd,
			[]int{},
			[]byte{byte(OpAdd)},
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

func TestInstructions_String(t *testing.T) {
	instructions := []Instructions{
		MakeInstruction(OpAdd),
		MakeInstruction(OpConstant, 2),
		MakeInstruction(OpConstant, 65535),
	}
	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
`

	concatenated := Instructions{}
	for _, ins := range instructions {
		concatenated = append(concatenated, ins...)
	}

	if concatenated.String() != expected {
		t.Errorf("wrongly formatted instructions. got = %q, want = %q", concatenated.String(), expected)
	}
}

func TestReadOperands(t *testing.T) {
	testCases := []struct {
		name      string
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{"OpConstant", OpConstant, []int{65535}, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instruction := MakeInstruction(tc.op, tc.operands...)

			def, err := Lookup(byte(tc.op))
			if err != nil {
				t.Fatalf("definition not found: %q", err)
			}

			operandsRead, n := ReadOperands(def, instruction[1:])
			if n != tc.bytesRead {
				t.Fatalf("n wrong. got = %d, want = %d", n, tc.bytesRead)
			}

			for i, want := range tc.operands {
				if operandsRead[i] != want {
					t.Errorf("operand wrong. got = %d, want = %d", operandsRead[i], want)
				}
			}
		})
	}
}
