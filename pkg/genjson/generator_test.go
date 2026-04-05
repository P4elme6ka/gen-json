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

	mustWrite(t, filepath.Join(tmp, "go.mod"), "module example.com/tmp\n\ngo 1.26\n")
	mustWrite(t, filepath.Join(tmp, "model.go"), `package sample

type Fancy string

type User struct {
	ID    int    `+"`json:\"id\"`"+`
	Name  string `+"`json:\"name\"`"+`
	Email string `+"`json:\"email,omitempty\"`"+`
	Nick  Fancy  `+"`json:\"nick,omitempty\"`"+`
}
`)

	cfg := Config{
		PackageDir:    tmp,
		Output:        filepath.Join(tmp, "zz_generated.genjson.go"),
		Types:         []string{"User"},
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

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
