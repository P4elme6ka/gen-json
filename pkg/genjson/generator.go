package genjson

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

func Generate(cfg Config) ([]byte, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	pkgName, types, err := loadTypes(cfg.PackageDir, cfg.Types)
	if err != nil {
		return nil, err
	}

	raw, err := render(pkgName, types, cfg)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(raw)
	if err != nil {
		return nil, fmt.Errorf("format generated code: %w\n--- raw ---\n%s", err, string(raw))
	}

	return formatted, nil
}

// Explain returns a human-friendly report for verbose output.
// It does not render code.
func Explain(cfg Config) (Report, error) {
	if err := cfg.Validate(); err != nil {
		return Report{}, err
	}
	pkgName, types, err := loadTypes(cfg.PackageDir, cfg.Types)
	if err != nil {
		return Report{}, err
	}
	return buildReport(pkgName, types, cfg), nil
}

func Write(cfg Config) (string, error) {
	code, err := Generate(cfg)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(cfg.Output, code, 0o644); err != nil {
		return "", fmt.Errorf("write output %q: %w", cfg.Output, err)
	}
	return cfg.Output, nil
}

func render(pkgName string, types []typeSpec, cfg Config) ([]byte, error) {
	tmpl, err := template.New("generated").Parse(generatedFileTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse code template: %w", err)
	}

	data := renderFileData{
		PackageName:           pkgName,
		EmitMarshaler:         cfg.EmitMarshaler,
		UnknownFields:         cfg.HasFeature(FeatureUnknownFields),
		RequiredFieldsEnabled: cfg.HasFeature(FeatureRequiredFields),
		Types:                 make([]renderType, 0, len(types)),
	}

	known := map[string]bool{}
	for _, t := range types {
		known[t.Name] = true
	}

	for _, t := range types {
		rt := renderType{
			Name:   t.Name,
			Fields: make([]renderField, 0, len(t.Fields)),
		}
		for _, f := range t.Fields {
			kind := f.DecodeKind
			baseType := f.BaseType
			elemKind := f.ElemKind
			elemType := f.ElemType
			// Only emit nested struct calls for types that are generated in this file.
			if kind == "struct" && !known[baseType] {
				kind, baseType = "json", ""
			}
			if kind == "ptr_struct" && !known[baseType] {
				kind, baseType = "json", ""
			}
			if kind == "slice" && elemKind == "struct" && !known[elemType] {
				kind, baseType, elemKind, elemType = "json", "", "", ""
			}
			if kind == "map" && elemKind == "struct" && !known[elemType] {
				kind, baseType, elemKind, elemType = "json", "", "", ""
			}
			rf := renderField{
				GoName:     f.GoName,
				JSONName:   f.JSONName,
				OmitEmpty:  f.OmitEmpty,
				Required:   cfg.HasFeature(FeatureRequiredFields) && !f.OmitEmpty,
				DecodeKind: kind,
				BaseType:   baseType,
				ElemKind:   elemKind,
				ElemType:   elemType,
			}
			rt.Fields = append(rt.Fields, rf)
			if rf.DecodeKind == "uuid" || rf.DecodeKind == "ptr_uuid" {
				data.UsesUUID = true
			}
			if rf.Required {
				rt.RequiredFields = append(rt.RequiredFields, rf)
			}
		}
		data.Types = append(data.Types, rt)
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("execute code template: %w", err)
	}
	return b.Bytes(), nil
}

type renderFileData struct {
	PackageName           string
	EmitMarshaler         bool
	UnknownFields         bool
	RequiredFieldsEnabled bool
	UsesUUID              bool
	Types                 []renderType
}

type renderType struct {
	Name           string
	Fields         []renderField
	RequiredFields []renderField
}

type renderField struct {
	GoName     string
	JSONName   string
	OmitEmpty  bool
	Required   bool
	DecodeKind string
	BaseType   string
	ElemKind   string
	ElemType   string
}

func buildReport(pkgName string, types []typeSpec, cfg Config) Report {
	r := Report{
		PackageName:   pkgName,
		PackageDir:    cfg.PackageDir,
		Output:        cfg.Output,
		Features:      append([]string(nil), cfg.Features...),
		EmitMarshaler: cfg.EmitMarshaler,
		Types:         make([]TypeReport, 0, len(types)),
	}

	for _, t := range types {
		tr := TypeReport{
			Name:       t.Name,
			FieldCount: len(t.Fields),
			Fields:     make([]FieldReport, 0, len(t.Fields)),
		}
		for _, f := range t.Fields {
			required := cfg.HasFeature(FeatureRequiredFields) && !f.OmitEmpty
			decodeFast := f.DecodeKind != "json"
			encodeFast := f.DecodeKind != "json"
			dPath := "slow"
			if decodeFast {
				dPath = "fast"
				tr.FastDecode++
			} else {
				tr.SlowDecode++
			}
			ePath := "slow"
			if encodeFast {
				ePath = "fast"
				tr.FastEncode++
			} else {
				tr.SlowEncode++
			}
			if required {
				tr.RequiredFields = append(tr.RequiredFields, f.JSONName)
			}
			tr.Fields = append(tr.Fields, FieldReport{
				GoName:     f.GoName,
				JSONName:   f.JSONName,
				OmitEmpty:  f.OmitEmpty,
				Required:   required,
				DecodeKind: f.DecodeKind,
				DecodePath: dPath,
				EncodePath: ePath,
			})
		}
		r.Types = append(r.Types, tr)
	}

	return r
}
