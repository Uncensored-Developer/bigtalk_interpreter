package evaluator

import (
	"BigTalk_Interpreter/object"
	"fmt"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(args ...object.IObject) object.IObject {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Items))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"tail": &object.Builtin{
		Fn: func(args ...object.IObject) object.IObject {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `tail` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Items)
			if length > 0 {
				newItems := make([]object.IObject, length-1)
				copy(newItems, arr.Items[1:length])
				return &object.Array{Items: newItems}
			}
			return NULL
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.IObject) object.IObject {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Items)

			newItems := make([]object.IObject, length+1)
			copy(newItems, arr.Items)
			newItems[length] = args[1]
			return &object.Array{Items: newItems}
		},
	},
	"print": &object.Builtin{
		Fn: func(args ...object.IObject) object.IObject {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
