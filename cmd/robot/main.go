package main

import (
	"fmt"

	_ "github.com/LogicHou/bftr/internal/store"
	"github.com/LogicHou/bftr/store/factory"
)

func main() {
	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
