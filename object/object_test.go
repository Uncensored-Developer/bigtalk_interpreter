package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	foo1 := &String{Value: "foo bar"}
	foo2 := &String{Value: "foo bar"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with thesame content have different hash keys")
	}

	if foo1.HashKey() != foo2.HashKey() {
		t.Errorf("strings with thesame content have different hash keys")
	}

	if hello1.HashKey() == foo1.HashKey() {
		t.Errorf("strings with different content have thesame hash keys")
	}
}
