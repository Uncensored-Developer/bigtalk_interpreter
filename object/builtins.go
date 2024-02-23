package object

import "fmt"

var BuiltinFunctions = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...IObject) IObject {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Items))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		}},
	},
	{
		"print",
		&Builtin{Fn: func(args ...IObject) IObject {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return nil
		}},
	},
	{
		"tail",
		&Builtin{Fn: func(args ...IObject) IObject {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `tail` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Items)
			if length > 0 {
				newItems := make([]IObject, length-1)
				copy(newItems, arr.Items[1:length])
				return &Array{Items: newItems}
			}
			return nil
		}},
	},
	{
		"push",
		&Builtin{Fn: func(args ...IObject) IObject {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Items)

			newItems := make([]IObject, length+1)
			copy(newItems, arr.Items)
			newItems[length] = args[1]
			return &Array{Items: newItems}
		}},
	},
}

func newError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinFunctionByName(name string) *Builtin {
	for _, def := range BuiltinFunctions {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
