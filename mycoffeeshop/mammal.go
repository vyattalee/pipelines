package main

import "fmt"

type Mammal interface {
	Say()
}

type Dog struct{}

type Cat struct{}

type Human struct{}

func (d Dog) Say() {
	fmt.Println("woof")
}

func (c Cat) Say() {
	fmt.Println("meow")
}

func (h Human) Say() {
	fmt.Println("speak")
}
