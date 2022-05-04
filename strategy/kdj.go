package strategy

import (
	"fmt"

	"github.com/LogicHou/bftr/datahandle"
)

type KDJ struct {
}

func NewStrategyKDJ() *KDJ {
	return &KDJ{}
}

func (this *KDJ) Run(client datahandle.Client, kline datahandle.Kline, trade datahandle.Trade) {
	fmt.Println("strategyKDJ run")
	client.Start()
}
