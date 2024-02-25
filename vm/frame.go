package vm

import (
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/object"
)

type Frame struct {
	closure     *object.Closure // Compiled function reference by the Frame
	ip          int             // Instruction pointer in this frame for this function
	basePointer int             // Points to the bottom of the stack of the current call frame
}

func NewFrame(closure *object.Closure, basePointer int) *Frame {
	return &Frame{closure: closure, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.closure.Fn.Instructions
}
