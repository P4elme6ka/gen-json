package main

import (
	"log"

	"gen-json/examples/basic"
)

func main() {
	if err := basic.Demo(); err != nil {
		log.Fatal(err)
	}
}
