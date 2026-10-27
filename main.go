package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

type stringSlice []string

func (s *stringSlice) String() string { return fmt.Sprintf("%v", *s) }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var (
	stdinData  []byte
	stdinErr   error
	stdinRead  bool
)

func readStdinOnce() ([]byte, error) {
	if stdinRead {
		return stdinData, stdinErr
	}
	stdinRead = true
	stdinData, stdinErr = io.ReadAll(os.Stdin)
	return stdinData, stdinErr
}

func resetStdinCache() {
	stdinData = nil
	stdinErr = nil
	stdinRead = false
}

func parseJSON(data []byte) (interface{}, error) {
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("empty JSON input")
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("trailing data after top-level JSON value")
		}
		return nil, err
	}
	return v, nil
}

func readJSON(path string) (interface{}, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = readStdinOnce()
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}
	return parseJSON(data)
}

func main() {
	var ignore stringSlice
	var quiet bool
	var output string
	colorFlag := flag.String("color", "auto", "color output: auto, always, never")
	flag.StringVar(&output, "output", "text", "output format: text or json")
	flag.StringVar(&output, "o", "text", "alias for --output")
	flag.StringVar(&output, "format", "text", "alias for --output")
	flag.BoolVar(&quiet, "q", false, "quiet (print nothing, only exit status)")
	flag.BoolVar(&quiet, "quiet", false, "alias for -q")
	flag.Var(&ignore, "ignore", "key to ignore (repeatable)")
	flag.Var(&ignore, "i", "alias for --ignore")
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

	switch output {
	case "text", "json":
	default:
		fmt.Fprintf(os.Stderr, "invalid --output value: %s\n", output)
		os.Exit(2)
	}

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

	if !quiet {
		switch output {
		case "json":
			out := changes
			if out == nil {
				out = []Change{}
			}
			encoded, err := json.Marshal(out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error encoding output: %v\n", err)
				os.Exit(2)
			}
			fmt.Println(string(encoded))
		default:
			for _, c := range changes {
				fmt.Println(formatChange(c, useColor))
			}
		}
	}
	if len(changes) > 0 {
		os.Exit(1)
	}
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
