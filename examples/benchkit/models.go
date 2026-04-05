package benchkit

// This package exists to provide an easily extensible benchmark fixture set.
// Add new structs here and extend the data table in benchmark_test.go.

import "github.com/google/uuid"

type A struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type B struct {
	OK    bool    `json:"ok"`
	Count uint64  `json:"count"`
	Rate  float64 `json:"rate"`
}

type C struct {
	Title string `json:"title"`
	Tag1  string `json:"tag1"`
	Tag2  string `json:"tag2"`
	N1    int    `json:"n1"`
	N2    int    `json:"n2"`
	N3    int    `json:"n3"`
}

type D struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type E struct {
	Optional *string `json:"optional,omitempty"`
	N        *int    `json:"n,omitempty"`
	Flag     *bool   `json:"flag,omitempty"`
}

// Complex: arrays, nested structs, maps, and UUID.
type F struct {
	IDs      []uuid.UUID          `json:"ids"`
	ByID     map[string]uuid.UUID `json:"by_id"`
	Counters map[string]uint64    `json:"counters"`
}

type GItem struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

type G struct {
	User  A                `json:"user"`
	Items []GItem          `json:"items"`
	Index map[string]GItem `json:"index"`
}

type H struct {
	Root   uuid.UUID `json:"root"`
	Tree   HTree     `json:"tree"`
	Levels [][]int   `json:"levels"`
}

type HTree struct {
	Left  *uuid.UUID `json:"left,omitempty"`
	Right *uuid.UUID `json:"right,omitempty"`
}
