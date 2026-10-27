package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestParseJSONValid(t *testing.T) {
	v, err := parseJSON([]byte(`{"a": 1}`))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := v.(map[string]interface{})
	if !ok || got["a"].(float64) != 1 {
		t.Fatalf("unexpected value: %#v", v)
	}
}

func TestParseJSONTrailingNewline(t *testing.T) {
	v, err := parseJSON([]byte("1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := v.(float64); !ok || n != 1 {
		t.Fatalf("unexpected value: %#v", v)
	}
}

func TestParseJSONBOM(t *testing.T) {
	v, err := parseJSON(append([]byte{0xEF, 0xBB, 0xBF}, []byte(`"hello"`)...))
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := v.(string); !ok || s != "hello" {
		t.Fatalf("unexpected value: %#v", v)
	}
}

func TestParseJSONEmpty(t *testing.T) {
	if _, err := parseJSON([]byte("   ")); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseJSONTrailingData(t *testing.T) {
	if _, err := parseJSON([]byte(`1 2`)); err == nil {
		t.Fatal("expected error for trailing data")
	}
}

func TestChangeMarshalJSONAdd(t *testing.T) {
	c := Change{Path: "$.x", Kind: KindAdd, New: 0.0}
	encoded, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	if strings.Contains(text, `"old"`) {
		t.Fatalf("old should be omitted: %s", text)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got["new"].(float64) != 0 || got["path"] != "$.x" || got["kind"] != "add" {
		t.Fatalf("unexpected JSON: %v", got)
	}
}

func TestChangeMarshalJSONRemove(t *testing.T) {
	c := Change{Path: "$.x", Kind: KindRemove, Old: ""}
	encoded, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	if strings.Contains(text, `"new"`) {
		t.Fatalf("new should be omitted: %s", text)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got["old"].(string) != "" || got["path"] != "$.x" || got["kind"] != "remove" {
		t.Fatalf("unexpected JSON: %v", got)
	}
}

func TestChangeMarshalJSONChange(t *testing.T) {
	c := Change{Path: "$.x", Kind: KindChange, Old: false, New: 0.0}
	encoded, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["old"].(bool); !ok || got["old"].(bool) != false {
		t.Fatalf("old zero value should be preserved: %v", got)
	}
	if _, ok := got["new"].(float64); !ok || got["new"].(float64) != 0 {
		t.Fatalf("new zero value should be preserved: %v", got)
	}
	if !reflect.DeepEqual(got["path"], "$.x") {
		t.Fatalf("unexpected path: %v", got)
	}
}
