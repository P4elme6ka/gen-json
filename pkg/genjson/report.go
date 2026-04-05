package genjson

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Report describes what code paths will be generated.
// It is intended for verbose generator output.
type Report struct {
	PackageName   string
	PackageDir    string
	Output        string
	Features      []string
	EmitMarshaler bool
	Types         []TypeReport
}

type TypeReport struct {
	Name           string
	FieldCount     int
	FastDecode     int
	SlowDecode     int
	FastEncode     int
	SlowEncode     int
	RequiredFields []string
	Fields         []FieldReport
}

type FieldReport struct {
	GoName     string
	JSONName   string
	OmitEmpty  bool
	Required   bool
	DecodeKind string
	DecodePath string // "fast" or "slow"
	EncodePath string // "fast" or "slow"
}

func (r Report) WriteTo(w io.Writer) error {
	features := append([]string(nil), r.Features...)
	sort.Strings(features)

	if _, err := fmt.Fprintf(w, "genjson report\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  package: %s\n  dir:     %s\n  output:  %s\n", r.PackageName, r.PackageDir, r.Output); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  features: %s\n", strings.Join(features, ", ")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  emitMarshaler: %v\n", r.EmitMarshaler); err != nil {
		return err
	}

	for _, t := range r.Types {
		if _, err := fmt.Fprintf(w, "\n  type %s\n", t.Name); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    fields: %d\n", t.FieldCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    decode: fast=%d slow=%d\n", t.FastDecode, t.SlowDecode); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    encode: fast=%d slow=%d\n", t.FastEncode, t.SlowEncode); err != nil {
			return err
		}
		if len(t.RequiredFields) > 0 {
			if _, err := fmt.Fprintf(w, "    required: %s\n", strings.Join(t.RequiredFields, ", ")); err != nil {
				return err
			}
		}
		for _, f := range t.Fields {
			if _, err := fmt.Fprintf(w, "    - %s (json=%q kind=%s omitempty=%v required=%v) decode=%s encode=%s\n",
				f.GoName, f.JSONName, f.DecodeKind, f.OmitEmpty, f.Required, f.DecodePath, f.EncodePath,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
