package genjson

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"strconv"
	"strings"
)

type fieldSpec struct {
	GoName     string
	JSONName   string
	OmitEmpty  bool
	DecodeKind string
	BaseType   string
	ElemKind   string
	ElemType   string
}

type typeSpec struct {
	Name   string
	Fields []fieldSpec
}

func loadTypes(packageDir string, wanted []string) (string, []typeSpec, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, packageDir, func(info fs.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return "", nil, fmt.Errorf("parse package %q: %w", packageDir, err)
	}

	if len(pkgs) == 0 {
		return "", nil, fmt.Errorf("no Go package found in %q", packageDir)
	}

	var pkgName string
	var pkg *ast.Package
	for name, p := range pkgs {
		pkgName = name
		pkg = p
		break
	}

	wantMap := make(map[string]bool, len(wanted))
	for _, t := range wanted {
		wantMap[t] = true
	}
	knownStructs := wantMap

	found := map[string]typeSpec{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !wantMap[ts.Name.Name] {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return "", nil, fmt.Errorf("type %q is not a struct", ts.Name.Name)
				}
				fields, err := collectFields(st, knownStructs)
				if err != nil {
					return "", nil, fmt.Errorf("type %q: %w", ts.Name.Name, err)
				}
				found[ts.Name.Name] = typeSpec{Name: ts.Name.Name, Fields: fields}
			}
		}
	}

	out := make([]typeSpec, 0, len(wanted))
	for _, t := range wanted {
		ts, ok := found[t]
		if !ok {
			return "", nil, fmt.Errorf("type %q not found in package %q", t, packageDir)
		}
		out = append(out, ts)
	}

	return pkgName, out, nil
}

func collectFields(st *ast.StructType, knownStructs map[string]bool) ([]fieldSpec, error) {
	if st.Fields == nil {
		return nil, nil
	}

	fields := make([]fieldSpec, 0, len(st.Fields.List))
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			// v0/v1: embedded fields are intentionally unsupported to keep generation predictable.
			continue
		}

		jsonName, omitEmpty, skip, err := parseJSONTag(f.Tag)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		for _, name := range f.Names {
			if !name.IsExported() {
				continue
			}
			decodeKind, baseType, elemKind, elemType := classifyDecodeKind(f.Type, knownStructs)
			fieldJSONName := jsonName
			if fieldJSONName == "" {
				fieldJSONName = name.Name
			}
			fields = append(fields, fieldSpec{
				GoName:     name.Name,
				JSONName:   fieldJSONName,
				OmitEmpty:  omitEmpty,
				DecodeKind: decodeKind,
				BaseType:   baseType,
				ElemKind:   elemKind,
				ElemType:   elemType,
			})
		}
	}
	return fields, nil
}

func classifyDecodeKind(expr ast.Expr, knownStructs map[string]bool) (decodeKind string, baseType string, elemKind string, elemType string) {
	// Special cases for well-known external types.
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok {
			// github.com/google/uuid.UUID
			if pkg.Name == "uuid" && sel.Sel.Name == "UUID" {
				return "uuid", "uuid.UUID", "", ""
			}
		}
	}

	if ident, ok := expr.(*ast.Ident); ok {
		switch ident.Name {
		case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "byte", "rune":
			return ident.Name, ident.Name, "", ""
		default:
			if knownStructs != nil && knownStructs[ident.Name] {
				return "struct", ident.Name, "", ""
			}
			return "json", "", "", ""
		}
	}

	// []T
	if arr, ok := expr.(*ast.ArrayType); ok {
		// Only slices (len == nil) supported.
		if arr.Len != nil {
			return "json", "", "", ""
		}
		k, bt, ek, et := classifyDecodeKind(arr.Elt, knownStructs)
		// We only support slices of known kinds; else fallback to json.
		switch k {
		case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "byte", "rune", "uuid", "struct":
			return "slice", "", k, bt
		case "slice":
			// nested slice: represent as slice of slice.
			// ElemKind is "slice" and ElemType holds the inner element type (bt), while bt itself is unused.
			_ = ek
			return "slice", "", "slice", et
		default:
			return "json", "", "", ""
		}
	}

	// map[string]T
	if mp, ok := expr.(*ast.MapType); ok {
		keyIdent, ok := mp.Key.(*ast.Ident)
		if !ok || keyIdent.Name != "string" {
			return "json", "", "", ""
		}
		k, bt, _, _ := classifyDecodeKind(mp.Value, knownStructs)
		switch k {
		case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "byte", "rune", "uuid", "struct":
			return "map", "", k, bt
		default:
			return "json", "", "", ""
		}
	}

	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return "json", "", "", ""
	}

	ident, ok := star.X.(*ast.Ident)
	if !ok {
		// pointer to selector type, e.g. *uuid.UUID
		if sel, ok := star.X.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok {
				if pkg.Name == "uuid" && sel.Sel.Name == "UUID" {
					return "ptr_uuid", "uuid.UUID", "", ""
				}
			}
		}
		return "json", "", "", ""
	}

	switch ident.Name {
	case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "byte", "rune":
		return "ptr_" + ident.Name, ident.Name, "", ""
	default:
		if knownStructs != nil && knownStructs[ident.Name] {
			return "ptr_struct", ident.Name, "", ""
		}
		return "json", "", "", ""
	}
}

func parseJSONTag(tag *ast.BasicLit) (jsonName string, omitEmpty bool, skip bool, err error) {
	if tag == nil {
		return "", false, false, nil
	}

	raw, err := strconv.Unquote(tag.Value)
	if err != nil {
		return "", false, false, fmt.Errorf("invalid struct tag %q: %w", tag.Value, err)
	}

	v := reflect.StructTag(raw).Get("json")
	if v == "" {
		return "", false, false, nil
	}
	parts := strings.Split(v, ",")
	if len(parts) > 0 {
		if parts[0] == "-" {
			return "", false, true, nil
		}
		jsonName = parts[0]
	}
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitEmpty = true
		}
	}
	return jsonName, omitEmpty, false, nil
}
