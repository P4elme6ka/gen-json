package genjson

import (
	"fmt"
)

const (
	FeatureUnknownFields  = "unknown_fields"
	FeatureRequiredFields = "required_fields"
)

// Config controls what code is generated.
type Config struct {
	PackageDir    string
	Output        string
	Types         []string
	Features      []string
	EmitMarshaler bool
}

func (c Config) Validate() error {
	if c.PackageDir == "" {
		return fmt.Errorf("packageDir is required")
	}
	if len(c.Types) == 0 {
		return fmt.Errorf("types is required and must include at least one struct type")
	}
	if c.Output == "" {
		return fmt.Errorf("output is required")
	}

	seen := map[string]bool{}
	for _, f := range c.Features {
		if seen[f] {
			continue
		}
		seen[f] = true
		switch f {
		case FeatureUnknownFields, FeatureRequiredFields:
		default:
			return fmt.Errorf("unsupported feature %q", f)
		}
	}

	return nil
}

func (c Config) HasFeature(feature string) bool {
	for _, f := range c.Features {
		if f == feature {
			return true
		}
	}
	return false
}
