package genjson

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAndCompile(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Clean(filepath.Join(mustWD(t), "..", ".."))

	mustWrite(t, filepath.Join(tmp, "go.mod"), "module example.com/tmp\n\ngo 1.26\n\nrequire github.com/P4elme6ka/gen-json v0.0.0\n\nreplace github.com/P4elme6ka/gen-json => "+repoRoot+"\n")
	mustWrite(t, filepath.Join(tmp, "model.go"), `package sample

import "github.com/google/uuid"

type Fancy string

type User struct {
	ID    int    `+"`json:\"id\"`"+`
	Name  string `+"`json:\"name\"`"+`
	Email string `+"`json:\"email,omitempty\"`"+`
	Nick  Fancy  `+"`json:\"nick,omitempty\"`"+`
}

type Team struct {
	IDs    []uuid.UUID            `+"`json:\"ids\"`"+`
	Lookup map[string]uuid.UUID   `+"`json:\"lookup\"`"+`
}
`)

	cfg := Config{
		PackageDir:    tmp,
		Output:        filepath.Join(tmp, "zz_generated.genjson.go"),
		Types:         []string{"User", "Team"},
		Features:      []string{FeatureUnknownFields, FeatureRequiredFields},
		EmitMarshaler: true,
	}

	code, err := Generate(cfg)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(string(code), "func DecodeUser") {
		t.Fatalf("generated code misses DecodeUser")
	}
	if !strings.Contains(string(code), "UnknownFieldError") {
		t.Fatalf("generated code misses UnknownFieldError")
	}
	if !strings.Contains(string(code), "github.com/google/uuid") {
		t.Fatalf("generated code misses uuid import for nested UUID fields")
	}
	if !strings.Contains(string(code), "github.com/P4elme6ka/gen-json/pkg/genjson/runtime") {
		t.Fatalf("generated code misses runtime helper import")
	}
	if !strings.Contains(string(code), "func (v User) MarshalJSON") {
		t.Fatalf("generated code misses MarshalJSON implementation")
	}

	mustWrite(t, cfg.Output, string(code))

	cmd := exec.Command("go", "test", ".")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GOWORK=off")
	// Ensure deps needed by the generated code are available.
	get := exec.Command("go", "get", "github.com/google/uuid@v1.6.0")
	get.Dir = tmp
	get.Env = cmd.Env
	if out, err := get.CombinedOutput(); err != nil {
		t.Fatalf("go get uuid failed: %v\n%s", err, out)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\n%s", err, out)
	}
}

func mustWD(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %v", err)
	}
	return wd
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
