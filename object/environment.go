package object

type Environment struct {
	store map[string]IObject
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]IObject)
	return &Environment{store: s}
}

func NewWrappedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (IObject, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val IObject) IObject {
	e.store[name] = val
	return val
}
