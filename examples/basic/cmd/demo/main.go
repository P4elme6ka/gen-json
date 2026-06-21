package main

import (
	"log"

	"github.com/P4elme6ka/gen-json/examples/basic"
)

func main() {
	if err := basic.Demo(); err != nil {
		log.Fatal(err)
	}
}
