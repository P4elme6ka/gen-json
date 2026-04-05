package basic

import (
	"encoding/json"
	"testing"
)

type stdUserPlain struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	benchInputJSONPlain = []byte(`{"id":42,"name":"Artem","email":"artem@example.com"}`)
	benchStdUserPlain   = stdUserPlain{ID: 42, Name: "Artem", Email: "artem@example.com"}
	benchOurUserPlain   = UserPlain{ID: 42, Name: "Artem", Email: "artem@example.com"}

	sinkStdUserPlain stdUserPlain
	sinkOurUserPlain UserPlain
	_                = sinkStdUserPlain
	_                = sinkOurUserPlain
)

func BenchmarkDecodeStdJSONPlain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var out stdUserPlain
		if err := json.Unmarshal(benchInputJSONPlain, &out); err != nil {
			b.Fatal(err)
		}
		sinkStdUserPlain = out
	}
}

func BenchmarkDecodeGenJSONPlain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := DecodeUserPlain(benchInputJSONPlain)
		if err != nil {
			b.Fatal(err)
		}
		sinkOurUserPlain = out
	}
}

func BenchmarkEncodeStdJSONPlain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := json.Marshal(benchStdUserPlain)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}

func BenchmarkEncodeGenJSONPlain(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := EncodeUserPlain(benchOurUserPlain)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}
