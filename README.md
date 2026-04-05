# gen-json

`gen-json` (CLI: `genjson`) generates fast, type-specific JSON **encoders/decoders** for Go structs.

The generated code:
- avoids reflection
- uses tight byte-slice appends for encoding
- parses JSON with a small purpose-built decoder
- can optionally enforce required fields and/or reject unknown fields

> Output file name defaults to `zz_generated.genjson.go`.

## Install

```bash
go install ./cmd/genjson
```

## Quick start

### 1) Add some models

```go
// models.go
package demo

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}
```

### 2) Generate code

From the package directory:

```bash
genjson -dir . -types User -out zz_generated.genjson.go -emit-marshaler
```

This generates:
- `func DecodeUser([]byte) (User, error)`
- `func EncodeUser(User) ([]byte, error)`
- (optional) `User.MarshalJSON()` / `(*User).UnmarshalJSON([]byte)` when `-emit-marshaler` is set

### 3) Use it

```go
u := User{ID: 1, Name: "Artem", Email: "a@example.com"}

b, err := EncodeUser(u)
if err != nil {
	panic(err)
}

u2, err := DecodeUser(b)
if err != nil {
	panic(err)
}
_ = u2
```

## CLI

```text
genjson \
  -dir <package_dir> \
  -types <T1,T2,...> \
  -out <output.go> \
  [-features <unknown_fields,required_fields>] \
  [-emit-marshaler] \
  [-v]
```

Flags:
- `-dir`: target package directory (default: `.`)
- `-types`: comma-separated list of struct type names to generate
- `-out`: output file path (default: `zz_generated.genjson.go`)
- `-features`: comma-separated list of optional features:
  - `unknown_fields`: reject JSON objects containing unknown fields
  - `required_fields`: require all **non-omitempty** fields to be present in JSON
- `-emit-marshaler`: also emit `MarshalJSON` / `UnmarshalJSON` methods
- `-v`: print a human-readable report (kinds, fast/slow paths, required fields)

## Go API (library)

You can use the generator programmatically via `pkg/genjson`:

```go
import "gen-json/pkg/genjson"

cfg := genjson.Config{
	PackageDir: ".",
	Output:     "zz_generated.genjson.go",
	Types:      []string{"User", "Order"},
	Features:   []string{genjson.FeatureUnknownFields, genjson.FeatureRequiredFields},
	EmitMarshaler: true,
}

// Generate returns formatted Go source.
code, err := genjson.Generate(cfg)

// Write generates and writes cfg.Output.
outPath, err := genjson.Write(cfg)

// Explain returns a report without rendering code.
r, err := genjson.Explain(cfg)
```

## Supported field shapes

The generator recognizes common Go field types and chooses a **fast** encode/decode path for them.

### Primitives
- `string`, `*string` (supports `omitempty`)
- `bool`, `*bool`
- Signed ints: `int`, `int8`, `int16`, `int32`, `int64`, `rune` + pointer variants
- Unsigned ints: `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `byte` + pointer variants
- `float32`, `float64` + pointer variants (rejects NaN/Inf on encode)

### Structs
- Nested struct values (`T`) where `T` is also generated in the same output file

### Slices
- `[]T` where `T` is one of:
  - numbers / bool / string
  - `uuid`-like types implementing `encoding.TextMarshaler`
  - generated structs
- `[][]<number>` (2D numeric slices) are supported (used by bench suite)

### Maps
- `map[string]T` where `T` matches the same set as slice elements

### Fallback (slow path)
If a field type isn’t recognized for a fast path, the generator emits a fallback:
- encode: try `json.Marshaler` (`MarshalJSON`)
- decode: try `json.Unmarshaler` (`UnmarshalJSON`)

If neither interface is implemented, the generated code returns `UnsupportedFieldTypeError`.

## JSON tags

- Field name comes from `json:"name"`.
- `omitempty` is respected.
- `json:"-"` fields are ignored.

## Features

### `unknown_fields`
When enabled, decoding returns `UnknownFieldError` if the input contains a field not present in the struct.

### `required_fields`
When enabled, decoding returns `MissingFieldError` if a **non-omitempty** field is not present in the input.

Notes:
- `omitempty` fields are never required.
- A required field may still be decoded as `null` when its type is a pointer.

## Output

The generated file contains:
- `Encode<Type>` and `Decode<Type>` functions
- helper functions (small JSON decoder + encoder helpers)
- optional `MarshalJSON` / `UnmarshalJSON` methods

## Examples

- `examples/basic`: minimal end-to-end example
- `examples/benchkit`: benchmark suite comparing generated code to `encoding/json`

To run the bench suite:

```bash
go test ./examples/benchkit -run '^$' -bench . -benchmem
```

## License

See repository for license information.

