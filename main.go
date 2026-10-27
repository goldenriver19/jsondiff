package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprintf("%v", *s) }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var ignore stringSlice
	colorFlag := flag.String("color", "auto", "color output: auto, always, never")
	quiet := flag.Bool("q", false, "quiet (print nothing, only exit status)")
	flag.Var(&ignore, "ignore", "key to ignore (repeatable)")
	flag.Var(&ignore, "i", "alias for -ignore")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file1> <file2>\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Compare two JSON documents and report structural differences.")
		fmt.Fprintln(os.Stderr, "Use - as a file argument to read from stdin.")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(2)
	}

	a, err := readJSON(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", args[0], err)
		os.Exit(2)
	}
	b, err := readJSON(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", args[1], err)
		os.Exit(2)
	}

	changes := Diff(a, b, ignore)

	useColor := false
	switch *colorFlag {
	case "always":
		useColor = true
	case "never":
		useColor = false
	case "auto":
		if fi, err := os.Stdout.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
			useColor = true
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid --color value: %s\n", *colorFlag)
		os.Exit(2)
	}

	if !*quiet {
		for _, c := range changes {
			fmt.Println(formatChange(c, useColor))
		}
	}
	if len(changes) > 0 {
		os.Exit(1)
	}
}

func readJSON(path string) (interface{}, error) {
	var f *os.File
	var err error
	if path == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}
	dec := json.NewDecoder(f)
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func formatChange(c Change, color bool) string {
	prefix := ""
	value := ""
	switch c.Kind {
	case KindAdd:
		prefix = "+"
		if color {
			prefix = "\x1b[32m+\x1b[0m"
		}
		value = marshal(c.New)
	case KindRemove:
		prefix = "-"
		if color {
			prefix = "\x1b[31m-\x1b[0m"
		}
		value = marshal(c.Old)
	case KindChange:
		prefix = "~"
		if color {
			prefix = "\x1b[33m~\x1b[0m"
		}
		value = marshal(c.Old) + " -> " + marshal(c.New)
	}
	return fmt.Sprintf("%s %s: %s", prefix, c.Path, value)
}

func marshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
