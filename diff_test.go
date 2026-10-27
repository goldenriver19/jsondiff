package main

import (
	"reflect"
	"testing"
)

func TestEqual(t *testing.T) {
	a := map[string]interface{}{"x": 1.0, "y": []interface{}{1.0, 2.0}}
	b := map[string]interface{}{"x": 1.0, "y": []interface{}{1.0, 2.0}}
	if len(Diff(a, b, nil)) != 0 {
		t.Fatal("expected no changes")
	}
}

func TestAddRemoveChange(t *testing.T) {
	a := map[string]interface{}{"a": 1.0, "b": 2.0}
	b := map[string]interface{}{"b": 3.0, "c": 4.0}
	ch := Diff(a, b, nil)
	kinds := map[string]string{}
	for _, c := range ch {
		kinds[c.Path] = string(c.Kind)
	}
	if kinds["$.a"] != "remove" {
		t.Errorf("$.a: %v", kinds)
	}
	if kinds["$.b"] != "change" {
		t.Errorf("$.b: %v", kinds)
	}
	if kinds["$.c"] != "add" {
		t.Errorf("$.c: %v", kinds)
	}
}

func TestNestedPath(t *testing.T) {
	a := map[string]interface{}{"user": map[string]interface{}{"name": "alice"}}
	b := map[string]interface{}{"user": map[string]interface{}{"name": "bob"}}
	ch := Diff(a, b, nil)
	if len(ch) != 1 || ch[0].Path != "$.user.name" {
		t.Fatalf("%+v", ch)
	}
}

func TestArrayIndex(t *testing.T) {
	a := map[string]interface{}{"items": []interface{}{1.0, 2.0}}
	b := map[string]interface{}{"items": []interface{}{1.0, 3.0, 4.0}}
	ch := Diff(a, b, nil)
	paths := []string{}
	for _, c := range ch {
		paths = append(paths, c.Path)
	}
	if !reflect.DeepEqual(paths, []string{"$.items[1]", "$.items[2]"}) {
		t.Fatalf("%v", paths)
	}
}

func TestIgnore(t *testing.T) {
	a := map[string]interface{}{"secret": "old", "x": 1.0}
	b := map[string]interface{}{"secret": "new", "x": 2.0}
	ch := Diff(a, b, []string{"secret"})
	if len(ch) != 1 || ch[0].Path != "$.x" {
		t.Fatalf("%+v", ch)
	}
}

func TestRootScalar(t *testing.T) {
	ch := Diff(1.0, 2.0, nil)
	if len(ch) != 1 || ch[0].Path != "$" {
		t.Fatalf("%+v", ch)
	}
}
