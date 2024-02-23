package evaluator

import (
	"BigTalk_Interpreter/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltinFunctionByName("len"),
	"tail":  object.GetBuiltinFunctionByName("tail"),
	"push":  object.GetBuiltinFunctionByName("push"),
	"print": object.GetBuiltinFunctionByName("print"),
}
