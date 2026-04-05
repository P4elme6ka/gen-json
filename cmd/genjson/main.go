package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gen-json/pkg/genjson"
)

func main() {
	pkgDir := flag.String("dir", ".", "Target package directory")
	out := flag.String("out", "zz_generated.genjson.go", "Output file path")
	types := flag.String("types", "", "Comma-separated list of struct types to generate")
	features := flag.String("features", "", "Comma-separated list of features (unknown_fields, required_fields)")
	emitMarshaler := flag.Bool("emit-marshaler", false, "Emit MarshalJSON/UnmarshalJSON methods")
	verbose := flag.Bool("v", false, "Verbose: print a generation report")
	flag.Parse()

	cfg, err := parseFlags(*pkgDir, *out, *types, *features, *emitMarshaler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "genjson: %v\n", err)
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(2)
	}

	if *verbose {
		r, err := genjson.Explain(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "genjson: %v\n", err)
			os.Exit(1)
		}
		_ = r.WriteTo(os.Stderr)
	}

	outputPath, err := genjson.Write(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "genjson: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("generated %s\n", outputPath)
}

func parseFlags(pkgDir, outPath, typesCSV, featuresCSV string, emitMarshaler bool) (genjson.Config, error) {
	parseCSV := func(s string) []string {
		var out []string
		for _, part := range strings.Split(s, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			out = append(out, part)
		}
		return out
	}

	cfg := genjson.Config{
		PackageDir:    pkgDir,
		Output:        outPath,
		Types:         parseCSV(typesCSV),
		Features:      parseCSV(featuresCSV),
		EmitMarshaler: emitMarshaler,
	}
	if err := cfg.Validate(); err != nil {
		return genjson.Config{}, err
	}
	return cfg, nil
}
