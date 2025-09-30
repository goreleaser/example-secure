package main

import "testing"

func TestGetBytes(t *testing.T) {
	got, err := getBytes(Person{
		Name: "John",
		Age:  30,
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := "name: John\nage: 30\n"
	if expected != string(got) {
		t.Fatalf("expected %q but got %q", expected, string(got))
	}
}
