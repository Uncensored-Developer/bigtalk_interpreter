package compiler

import "testing"

func TestSymbolTable_Define(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("a=%+v, want =  %+v", a, expected["a"])
	}

	b := global.Define("b")
	if a != expected["a"] {
		t.Errorf("b=%+v, want =  %+v", b, expected["b"])
	}
}

func TestSymbolTable_Resolve_Global(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, symb := range expected {
		result, ok := global.Resolve(symb.Name)
		if !ok {
			t.Errorf("name %s could not be resolved.", symb.Name)
			continue
		}

		if result != symb {
			t.Errorf("expected %s to resolve to %+v, got = %+v", symb.Name, symb, result)
		}
	}
}
