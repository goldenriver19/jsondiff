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

Options:

- `-i` / `--ignore KEY` ignore a key everywhere (repeatable)
- `-q` / `--quiet` quiet mode (no output, only exit status)
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

## Tests

```
go test ./...
```
