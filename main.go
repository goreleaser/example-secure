package main

import (
	"fmt"
	"log"

	"go.yaml.in/yaml/v4"
)

func main() {
	bts, err := yaml.Marshal(struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  30,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\nyaml:\n\n%s\n", bts)
}
