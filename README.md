# jsondiff

A small command-line tool written in Go that compares two JSON documents and
reports structural differences (added, removed, and changed values) annotated
with JSON paths. Useful in scripts, CI checks, and debugging configuration
changes.

## Build

```
go build -o jsondiff .
```

## Usage

```
jsondiff [options] <file1> <file2>
```

Use `-` as one of the file arguments to read that document from standard input.
Both arguments may be `-`; standard input is then read once and used for both
documents.

Options:

- `-i` / `--ignore KEY` ignore a key everywhere (repeatable)
- `-q` / `--quiet` quiet mode (no output, only exit status)
- `-o` / `--output text|json` / `--format text|json` output format (default: text)
- `--color auto|always|never` control colored output (default: auto)

## Exit codes

- `0` documents are equivalent (after ignores)
- `1` differences were found
- `2` an error occurred (bad arguments, unreadable or invalid JSON)

## Example

`a.json`:

```json
{"name":"alice","age":30}
```

`b.json`:

```json
{"name":"bob","age":30,"city":"NYC"}
```

```
$ jsondiff a.json b.json
~ $.name: "alice" -> "bob"
+ $.city: "NYC"
```

Machine-readable output:

```
$ jsondiff --output json a.json b.json
[{"kind":"change","new":"bob","old":"alice","path":"$.name"},{"kind":"add","new":"NYC","path":"$.city"}]
```

## Tests

```
go test ./...
```
