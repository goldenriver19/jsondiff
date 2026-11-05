package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// ChangeKind describes the type of difference.
type ChangeKind string

const (
	KindAdd    ChangeKind = "add"
	KindRemove ChangeKind = "remove"
	KindChange ChangeKind = "change"
)

// Change represents a single structural difference between two documents.
type Change struct {
	Path string
	Kind ChangeKind
	Old  interface{}
	New  interface{}
}

// MarshalJSON renders a Change in a machine-readable form. Absent values are
// omitted according to the change kind, including zero values such as 0,
// false, or the empty string.
func (c Change) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"path": c.Path,
		"kind": c.Kind,
	}
	if c.Kind != KindAdd {
		obj["old"] = c.Old
	}
	if c.Kind != KindRemove {
		obj["new"] = c.New
	}
	return json.Marshal(obj)
}

// Diff compares two decoded JSON values and returns all differences.
// Keys listed in ignore are skipped entirely (including their subtrees).
func Diff(a, b interface{}, ignore []string) []Change {
	return DiffWithIgnores(a, b, ignore, nil)
}

// DiffWithIgnores compares two decoded JSON values and returns all
// differences. Keys listed in ignoreKeys are skipped everywhere, while paths
// listed in ignorePaths are skipped only at those exact JSON paths.
func DiffWithIgnores(a, b interface{}, ignoreKeys, ignorePaths []string) []Change {
	errIgnoredKeys := make(map[string]bool, len(ignoreKeys))
	for _, k := range ignoreKeys {
		errIgnoredKeys[k] = true
	}
	ignoredPaths := make(map[string]bool, len(ignorePaths))
	for _, p := range ignorePaths {
		ignoredPaths[p] = true
	}
	var changes []Change
	diffValue("$", a, b, errIgnoredKeys, ignoredPaths, &changes)
	return changes
}

func diffValue(path string, a, b interface{}, ignoreKeys, ignorePaths map[string]bool, changes *[]Change) {
	if ignorePaths[path] {
		return
	}

	// Both objects: compare union of keys.
	ma, oka := a.(map[string]interface{})
	mb, okb := b.(map[string]interface{})
	if oka && okb {
		keys := make(map[string]bool)
		for k := range ma {
			keys[k] = true
		}
		for k := range mb {
			keys[k] = true
		}
		sorted := make([]string, 0, len(keys))
		for k := range keys {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)
		for _, k := range sorted {
			if ignoreKeys[k] {
				continue
			}
			va, ina := ma[k]
			vb, inb := mb[k]
			child := path + "." + k
			if ignorePaths[child] {
				continue
			}
			if !ina {
				*changes = append(*changes, Change{Path: child, Kind: KindAdd, New: vb})
			} else if !inb {
				*changes = append(*changes, Change{Path: child, Kind: KindRemove, Old: va})
			} else {
				diffValue(child, va, vb, ignoreKeys, ignorePaths, changes)
			}
		}
		return
	}

	// Both arrays: compare element-wise by index.
	la, oka := a.([]interface{})
	lb, okb := b.([]interface{})
	if oka && okb {
		n := len(la)
		if len(lb) > n {
			n = len(lb)
		}
		for i := 0; i < n; i++ {
			child := fmt.Sprintf("%s[%d]", path, i)
			if ignorePaths[child] {
				continue
			}
			if i >= len(la) {
				*changes = append(*changes, Change{Path: child, Kind: KindAdd, New: lb[i]})
			} else if i >= len(lb) {
				*changes = append(*changes, Change{Path: child, Kind: KindRemove, Old: la[i]})
			} else {
				diffValue(child, la[i], lb[i], ignoreKeys, ignorePaths, changes)
			}
		}
		return
	}

	// Leaf: report a change when values differ.
	if !reflect.DeepEqual(a, b) {
		*changes = append(*changes, Change{Path: path, Kind: KindChange, Old: a, New: b})
	}
}
