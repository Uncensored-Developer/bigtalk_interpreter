package vm

import (
	"BigTalk_Interpreter/code"
	"BigTalk_Interpreter/object"
)

type Frame struct {
	fn *object.CompiledFunction // Compiled function reference by the Frame
	ip int                      // Instruction pointer in this frame for this function
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{fn: fn, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
