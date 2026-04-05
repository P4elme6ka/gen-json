package benchkit

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// To extend this suite:
//  1) Add a new struct type in models.go
//  2) Add it to //go:generate in gen.go
//  3) Add a new entry to the tables below

type codecCase[T any] struct {
	name   string
	json   []byte
	stdVal T
	decode func([]byte) (T, error)
	encode func(T) ([]byte, error)
}

func benchDecode[T any](b *testing.B, c codecCase[T]) {
	b.Helper()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := c.decode(c.json)
		if err != nil {
			b.Fatal(err)
		}
		_ = v
	}
}

func benchEncode[T any](b *testing.B, c codecCase[T]) {
	b.Helper()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		out, err := c.encode(c.stdVal)
		if err != nil {
			b.Fatal(err)
		}
		_ = out
	}
}

func BenchmarkDecodeStd(b *testing.B) {
	u1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	u2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	u3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	cases := []struct {
		name string
		json []byte
		fn   func([]byte) error
	}{
		{
			name: "A",
			json: []byte(`{"id":1,"name":"Artem","email":"a@example.com"}`),
			fn: func(data []byte) error {
				var v A
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "B",
			json: []byte(`{"ok":true,"count":123,"rate":3.14}`),
			fn: func(data []byte) error {
				var v B
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "C",
			json: []byte(`{"title":"hello","tag1":"a","tag2":"b","n1":1,"n2":2,"n3":3}`),
			fn: func(data []byte) error {
				var v C
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "D",
			json: []byte(`{"x":1,"y":2,"z":3}`),
			fn: func(data []byte) error {
				var v D
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "E",
			json: []byte(`{"optional":"x","n":5,"flag":true}`),
			fn: func(data []byte) error {
				var v E
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "F",
			json: []byte(`{"ids":["11111111-1111-1111-1111-111111111111","22222222-2222-2222-2222-222222222222"],"by_id":{"a":"33333333-3333-3333-3333-333333333333"},"counters":{"x":1,"y":2}}`),
			fn: func(data []byte) error {
				var v F
				_ = u1
				_ = u2
				_ = u3
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "G",
			json: []byte(`{"user":{"id":1,"name":"Artem","email":"a@example.com"},"items":[{"key":"a","value":1},{"key":"b","value":2}],"index":{"a":{"key":"a","value":1}}}`),
			fn: func(data []byte) error {
				var v G
				return json.Unmarshal(data, &v)
			},
		},
		{
			name: "H",
			json: []byte(`{"root":"11111111-1111-1111-1111-111111111111","tree":{"left":"22222222-2222-2222-2222-222222222222"},"levels":[[1,2],[3,4,5]]}`),
			fn: func(data []byte) error {
				var v H
				return json.Unmarshal(data, &v)
			},
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := c.fn(c.json); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeGen(b *testing.B) {
	cases := []struct {
		name string
		json []byte
		fn   func([]byte) error
	}{
		{
			name: "A",
			json: []byte(`{"id":1,"name":"Artem","email":"a@example.com"}`),
			fn: func(data []byte) error {
				_, err := DecodeA(data)
				return err
			},
		},
		{
			name: "B",
			json: []byte(`{"ok":true,"count":123,"rate":3.14}`),
			fn: func(data []byte) error {
				_, err := DecodeB(data)
				return err
			},
		},
		{
			name: "C",
			json: []byte(`{"title":"hello","tag1":"a","tag2":"b","n1":1,"n2":2,"n3":3}`),
			fn: func(data []byte) error {
				_, err := DecodeC(data)
				return err
			},
		},
		{
			name: "D",
			json: []byte(`{"x":1,"y":2,"z":3}`),
			fn: func(data []byte) error {
				_, err := DecodeD(data)
				return err
			},
		},
		{
			name: "E",
			json: []byte(`{"optional":"x","n":5,"flag":true}`),
			fn: func(data []byte) error {
				_, err := DecodeE(data)
				return err
			},
		},
		{
			name: "F",
			json: []byte(`{"ids":["11111111-1111-1111-1111-111111111111","22222222-2222-2222-2222-222222222222"],"by_id":{"a":"33333333-3333-3333-3333-333333333333"},"counters":{"x":1,"y":2}}`),
			fn: func(data []byte) error {
				_, err := DecodeF(data)
				return err
			},
		},
		{
			name: "G",
			json: []byte(`{"user":{"id":1,"name":"Artem","email":"a@example.com"},"items":[{"key":"a","value":1},{"key":"b","value":2}],"index":{"a":{"key":"a","value":1}}}`),
			fn: func(data []byte) error {
				_, err := DecodeG(data)
				return err
			},
		},
		{
			name: "H",
			json: []byte(`{"root":"11111111-1111-1111-1111-111111111111","tree":{"left":"22222222-2222-2222-2222-222222222222"},"levels":[[1,2],[3,4,5]]}`),
			fn: func(data []byte) error {
				_, err := DecodeH(data)
				return err
			},
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := c.fn(c.json); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodeStd(b *testing.B) {
	opt := "x"
	n := 5
	flag := true
	u1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	u2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	u3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	cases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "A",
			fn: func() error {
				_, err := json.Marshal(A{ID: 1, Name: "Artem", Email: "a@example.com"})
				return err
			},
		},
		{
			name: "B",
			fn: func() error {
				_, err := json.Marshal(B{OK: true, Count: 123, Rate: 3.14})
				return err
			},
		},
		{
			name: "C",
			fn: func() error {
				_, err := json.Marshal(C{Title: "hello", Tag1: "a", Tag2: "b", N1: 1, N2: 2, N3: 3})
				return err
			},
		},
		{
			name: "D",
			fn: func() error {
				_, err := json.Marshal(D{X: 1, Y: 2, Z: 3})
				return err
			},
		},
		{
			name: "E",
			fn: func() error {
				_, err := json.Marshal(E{Optional: &opt, N: &n, Flag: &flag})
				return err
			},
		},
		{
			name: "F",
			fn: func() error {
				_, err := json.Marshal(F{IDs: []uuid.UUID{u1, u2}, ByID: map[string]uuid.UUID{"a": u3}, Counters: map[string]uint64{"x": 1, "y": 2}})
				return err
			},
		},
		{
			name: "G",
			fn: func() error {
				_, err := json.Marshal(G{User: A{ID: 1, Name: "Artem", Email: "a@example.com"}, Items: []GItem{{Key: "a", Value: 1}, {Key: "b", Value: 2}}, Index: map[string]GItem{"a": {Key: "a", Value: 1}}})
				return err
			},
		},
		{
			name: "H",
			fn: func() error {
				v := H{Root: u1, Levels: [][]int{{1, 2}, {3, 4, 5}}}
				v.Tree.Left = &u2
				_, err := json.Marshal(v)
				return err
			},
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := c.fn(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodeGen(b *testing.B) {
	opt := "x"
	n := 5
	flag := true
	u1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	u2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	u3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	cases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "A",
			fn: func() error {
				_, err := EncodeA(A{ID: 1, Name: "Artem", Email: "a@example.com"})
				return err
			},
		},
		{
			name: "B",
			fn: func() error {
				_, err := EncodeB(B{OK: true, Count: 123, Rate: 3.14})
				return err
			},
		},
		{
			name: "C",
			fn: func() error {
				_, err := EncodeC(C{Title: "hello", Tag1: "a", Tag2: "b", N1: 1, N2: 2, N3: 3})
				return err
			},
		},
		{
			name: "D",
			fn: func() error {
				_, err := EncodeD(D{X: 1, Y: 2, Z: 3})
				return err
			},
		},
		{
			name: "E",
			fn: func() error {
				_, err := EncodeE(E{Optional: &opt, N: &n, Flag: &flag})
				return err
			},
		},
		{
			name: "F",
			fn: func() error {
				_, err := EncodeF(F{IDs: []uuid.UUID{u1, u2}, ByID: map[string]uuid.UUID{"a": u3}, Counters: map[string]uint64{"x": 1, "y": 2}})
				return err
			},
		},
		{
			name: "G",
			fn: func() error {
				_, err := EncodeG(G{User: A{ID: 1, Name: "Artem", Email: "a@example.com"}, Items: []GItem{{Key: "a", Value: 1}, {Key: "b", Value: 2}}, Index: map[string]GItem{"a": {Key: "a", Value: 1}}})
				return err
			},
		},
		{
			name: "H",
			fn: func() error {
				v := H{Root: u1, Levels: [][]int{{1, 2}, {3, 4, 5}}}
				v.Tree.Left = &u2
				_, err := EncodeH(v)
				return err
			},
		},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := c.fn(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
