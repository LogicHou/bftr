package main

import (
	"fmt"
	"time"

	_ "github.com/LogicHou/bftr/internal/store"
	"github.com/LogicHou/bftr/strategy"
	"github.com/LogicHou/bftr/workerpool"
)

func main() {
	p := workerpool.New(1, workerpool.WithPreAllocWorkers(true), workerpool.WithBlock(true))

	time.Sleep(2 * time.Second)
	err := p.Schedule(func() {
		strategyKDJ := strategy.NewStrategyKDJ()
		strategyKDJ.Run()
	})
	if err != nil {
		fmt.Printf("task kdj: error: %s\n", err.Error())
	}
	fmt.Println("workerpool start ok")

	p.Free()

}
