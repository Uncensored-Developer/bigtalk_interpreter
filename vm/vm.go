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
	MaxFrames   = 1024
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VirtualMachine struct {
	constants []object.IObject
	stack     []object.IObject
	sp        int // Points to the next value

	globals []object.IObject

	frames      []*Frame
	framesIndex int
}

func NewVirtualMachine(bytecode *compiler.ByteCode) *VirtualMachine {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VirtualMachine{
		constants:   bytecode.Constants,
		stack:       make([]object.IObject, StackSize),
		sp:          0,
		globals:     make([]object.IObject, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
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
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for v.currentFrame().ip < len(v.currentFrame().Instructions())-1 {
		v.currentFrame().ip++

		ip = v.currentFrame().ip
		ins = v.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			index := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2
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
			pos := int(code.ReadUint16(ins[ip+1:]))
			// Set the instruction pointer to the target of our jump
			// accounting for the default increment of i in the for-loop
			v.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			condition := v.pop()
			if !isTruthy(condition) {
				v.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2
			v.globals[globalIdx] = v.pop()
		case code.OpGetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			v.currentFrame().ip += 2
			err := v.push(v.globals[globalIdx])
			if err != nil {
				return err
			}
		case code.OpArray:
			arrayLength := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			array := v.buildArray(v.sp-arrayLength, v.sp)
			v.sp = v.sp - arrayLength

			err := v.push(array)
			if err != nil {
				return err
			}
		case code.OpMap:
			mapLength := int(code.ReadUint16(ins[ip+1:]))
			v.currentFrame().ip += 2

			mapObj, err := v.buildMap(v.sp-mapLength, v.sp)
			if err != nil {
				return err
			}
			v.sp = v.sp - mapLength

			err = v.push(mapObj)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := v.pop()
			obj := v.pop()

			err := v.executeIndexExpression(obj, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			argsCount := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			err := v.executeCall(int(argsCount))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			// first pop return value off the stack
			returnVal := v.pop()

			// pop the just executed frame of the frame stack
			frame := v.popFrame()
			v.sp = frame.basePointer - 1

			err := v.push(returnVal)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := v.popFrame()
			v.sp = frame.basePointer - 1

			err := v.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()
			v.stack[frame.basePointer+int(localIndex)] = v.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			frame := v.currentFrame()
			err := v.push(v.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			v.currentFrame().ip += 1

			definition := object.BuiltinFunctions[builtinIndex]
			err := v.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			_ = code.ReadUint8(ins[ip+3:])
			v.currentFrame().ip += 3

			err := v.pushClosure(int(constIndex))
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

func (v *VirtualMachine) executeIndexExpression(obj, index object.IObject) error {
	switch {
	case obj.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return v.executeArrayIndex(obj, index)
	case obj.Type() == object.MAP_OBJ:
		return v.executeMapIndex(obj, index)
	default:
		return fmt.Errorf("index operator not supported for %s", obj.Type())
	}
}

func (v *VirtualMachine) executeArrayIndex(array, index object.IObject) error {
	arrayObj := array.(*object.Array)
	i := index.(*object.Integer).Value
	maxLen := int64(len(arrayObj.Items) - 1)

	if i < 0 || i > maxLen {
		return v.push(Null)
	}

	return v.push(arrayObj.Items[i])
}

func (v *VirtualMachine) executeMapIndex(hash, index object.IObject) error {
	mapObj := hash.(*object.Map)

	key, ok := index.(object.IHashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := mapObj.Pairs[key.HashKey()]
	if !ok {
		return v.push(Null)
	}

	return v.push(pair.Value)
}

func (v *VirtualMachine) buildArray(startIndex, endIndex int) object.IObject {
	items := make([]object.IObject, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		items[i-startIndex] = v.stack[i]
	}
	return &object.Array{Items: items}
}

// buildMap constructs a new instance of object.Map using the elements from the stack within the specified range.
// It iterates over the stack starting from startIndex and ending at endIndex, by incrementing the index by 2 in each iteration.
// For every pair of stack elements at indices i and i+1, it creates a new object.MapPair with the key as the element at index i, and the value as the element at index i+1.
// It then checks if the key implements the object.IHashable interface. If not, it returns an error with a message indicating that the key is not usable as a hash key.
// Otherwise, it computes the hash key using the key's HashKey() method and adds the pair to the hashedPairs map using the hash key as the key.
// Finally, it returns a new instance of object.Map with the hashedPairs as the pairs field.
// If an error occurs during the construction of the map, it returns nil and the error.
func (v *VirtualMachine) buildMap(startIndex, endIndex int) (object.IObject, error) {
	hashedPairs := make(map[object.HashKey]object.MapPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := v.stack[i]
		value := v.stack[i+1]

		pair := object.MapPair{Key: key, Value: value}

		hashKey, ok := key.(object.IHashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		hashedPairs[hashKey.HashKey()] = pair
	}
	return &object.Map{Pairs: hashedPairs}, nil
}

func (v *VirtualMachine) currentFrame() *Frame {
	return v.frames[v.framesIndex-1]
}

func (v *VirtualMachine) pushFrame(f *Frame) {
	v.frames[v.framesIndex] = f
	v.framesIndex++
}

func (v *VirtualMachine) popFrame() *Frame {
	v.framesIndex--
	return v.frames[v.framesIndex]
}

func (v *VirtualMachine) callClosure(closure *object.Closure, argsCount int) error {
	if argsCount != closure.Fn.ParametersCount {
		return fmt.Errorf(
			"wrong number of arguments: got = %d, want = %d", closure.Fn.ParametersCount, argsCount)
	}

	frame := NewFrame(closure, v.sp-argsCount)
	v.pushFrame(frame)

	// Allocate space for the local bindings on the stack
	// by increasing the value of the stack pointer (sp)
	v.sp = frame.basePointer + closure.Fn.LocalsCount
	return nil
}

func (v *VirtualMachine) executeCall(argsCount int) error {
	called := v.stack[v.sp-1-argsCount]
	switch called := called.(type) {
	case *object.Closure:
		return v.callClosure(called, argsCount)
	case *object.Builtin:
		return v.callBuiltin(called, argsCount)
	default:
		return fmt.Errorf("calling a non-function or non-builtin")
	}
}

func (v *VirtualMachine) callBuiltin(builtin *object.Builtin, argsCount int) error {
	args := v.stack[v.sp-argsCount : v.sp]

	result := builtin.Fn(args...)
	v.sp = v.sp - argsCount - 1

	var err error
	if result != nil {
		err = v.push(result)
	} else {
		err = v.push(Null)
	}
	if err != nil {
		return err
	}
	return nil
}

func (v *VirtualMachine) pushClosure(constIndex int) error {
	constant := v.constants[constIndex]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("constant is not *object.CompiledFunction: %+v", constant)
	}

	closure := &object.Closure{Fn: fn}
	return v.push(closure)
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
