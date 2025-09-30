package main

import (
	"fmt"
	"log"

	"go.yaml.in/yaml/v4"
)

func main() {
	bts, err := getBytes(Person{
		Name: "John",
		Age:  30,
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\nyaml:\n\n%s\n", bts)
}

func getBytes(p Person) ([]byte, error) {
	bts, err := yaml.Marshal(p)
	return bts, err
}

type Person struct {
	Name string
	Age  int
}
