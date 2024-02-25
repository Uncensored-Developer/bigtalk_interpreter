package compiler

import "testing"

func TestSymbolTableShadowingFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	expected := Symbol{Name: "a", Scope: GlobalScope, Index: 0}

	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %q could not be resolved", expected.Name)
	}

	if result != expected {
		t.Errorf("expected %q to resolve to %+v, got=%+v", expected.Name, expected, result)
	}
}

func TestSymbolTableDefineAndResolveFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	expected := Symbol{Name: "a", Scope: FunctionScope, Index: 0}

	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %q could not be resolved", expected.Name)
	}

	if result != expected {
		t.Errorf("expected %q to resolve to %+v, got=%+v", expected.Name, expected, result)
	}
}

func TestSymbolTable_Resolve_UnresolvableFreeVariables(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewWrappedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewWrappedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %q could not be resolved", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %q to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}

	expectedUnresolvable := []string{
		"b",
		"d",
	}

	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %q resolved, but was expected not to", name)
		}
	}
}

func TestSymbolTable_Resolve_FreeVariables(t *testing.T) {
	// let a = 1;
	// let b = 2;
	//
	// let firstLocal = fn() {
	//	 let c = 3;
	//	 let d = 4;
	//	 a + b + c + d;
	//
	//	 let secondLocal = fn() {
	//		 let e = 5;
	//		 let f = 6;
	//		 a + b + c + d + e + f;
	//	 };
	// };

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
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tc := range testCases {
		for _, sym := range tc.expectedSymbols {
			result, ok := tc.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %q could not be resolved", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %q to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}

		if len(tc.table.FreeSymbols) != len(tc.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d",
				len(tc.table.FreeSymbols), len(tc.expectedFreeSymbols))
			continue
		}
		for i, sym := range tc.expectedFreeSymbols {
			result := tc.table.FreeSymbols[i]
			if result != sym {
				t.Errorf("wrong free symbol. got=%+v, want=%+v",
					result, sym)
			}
		}
	}
}

func TestSymbolTable_Define_Resole_Builtins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewWrappedSymbolTable(global)
	secondLocal := NewWrappedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "b", Scope: BuiltinScope, Index: 1},
		{Name: "c", Scope: BuiltinScope, Index: 2},
		{Name: "d", Scope: BuiltinScope, Index: 3},
	}

	for i, sym := range expected {
		global.DefineBuiltin(i, sym.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %q could not be resolved", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %q to resolve to %+v, got = %+v", sym.Name, sym, result)
			}
		}
	}
}

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
				t.Errorf("name %q could not be resolved", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %q to resolve to %+v, got = %+v", sym.Name, sym, result)
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
			t.Errorf("name %q could not be resolved", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %q to resolve to %+v, got = %+v", sym.Name, sym, result)
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
			t.Errorf("name %q could not be resolved.", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %q to resolve to %+v, got = %+v", sym.Name, sym, result)
		}
	}
}
