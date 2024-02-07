package object

type Environment struct {
	store map[string]IObject
}

func NewEnvironment() *Environment {
	s := make(map[string]IObject)
	return &Environment{store: s}
}

func (e *Environment) Get(name string) (IObject, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

func (e *Environment) Set(name string, val IObject) IObject {
	e.store[name] = val
	return val
}
