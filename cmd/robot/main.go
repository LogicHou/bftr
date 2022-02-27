package main

import (
	"fmt"

	"github.com/LogicHou/bftr/store/factory"
)

func main() {
	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
