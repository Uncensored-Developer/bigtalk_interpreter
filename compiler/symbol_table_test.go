package compiler

import "testing"

func TestSymbolTable_Resolve_NestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewWrappedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewWrappedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	testCases := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tc := range testCases {
		for _, sym := range tc.expectedSymbols {
			result, ok := tc.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s could not be resolved", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got = %+v", sym.Name, sym, result)
			}
		}
	}
}

func TestSymbolTable_Define(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
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

	firstLocal := NewWrappedSymbolTable(global)
	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("c=%+v, want =  %+v", c, expected["c"])
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("d=%+v, want =  %+v", d, expected["d"])
	}

	secondLocal := NewWrappedSymbolTable(firstLocal)
	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("e=%+v, want =  %+v", e, expected["e"])
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("f=%+v, want =  %+v", f, expected["f"])
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

	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s could not be resolved", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got = %+v", sym.Name, sym, result)
		}
	}
}

func TestSymbolTable_Resolve_Local(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewWrappedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s could not be resolved.", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got = %+v", sym.Name, sym, result)
		}
	}
}
