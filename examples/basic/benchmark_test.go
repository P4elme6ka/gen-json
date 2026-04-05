package basic

import (
	"encoding/json"
	"testing"
)

type stdUser struct {
	ID    int              `json:"id"`
	Name  string           `json:"name"`
	Email string           `json:"email"`
	Nick  LowerUpperString `json:"nick,omitempty"`
}

var (
	benchInputJSON = []byte(`{"id":42,"name":"Artem","email":"artem@example.com","nick":"miXeD"}`)
	benchStdUser   = stdUser{ID: 42, Name: "Artem", Email: "artem@example.com", Nick: "mixed"}
	benchOurUser   = User{ID: 42, Name: "Artem", Email: "artem@example.com", Nick: "mixed"}

	sinkStdUser stdUser
	sinkOurUser User
	sinkBytes   []byte
)

func BenchmarkDecodeStdJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var out stdUser
		if err := json.Unmarshal(benchInputJSON, &out); err != nil {
			b.Fatal(err)
		}
		sinkStdUser = out
	}
}

func BenchmarkDecodeGenJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := DecodeUser(benchInputJSON)
		if err != nil {
			b.Fatal(err)
		}
		sinkOurUser = out
	}
}

func BenchmarkEncodeStdJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := json.Marshal(benchStdUser)
		if err != nil {
			b.Fatal(err)
		}
		sinkBytes = out
	}
}

func BenchmarkEncodeGenJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := EncodeUser(benchOurUser)
		if err != nil {
			b.Fatal(err)
		}
		sinkBytes = out
	}
}
