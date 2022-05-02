package main

import (
	"fmt"
	"os"

	"github.com/LogicHou/bftr/datasource/binance"
	_ "github.com/LogicHou/bftr/internal/store"
	"github.com/LogicHou/bftr/store/factory"
)

func main() {
	KlineSrv := binance.KlineSrv{}
	err := KlineSrv.Get(145)
	if err != nil {
		fmt.Println(err)
		return
	}
	KlineSrv.WithKdj()
	for _, v := range KlineSrv.Klines {
		fmt.Printf("%.2f %.2f %.2f\n", v.Close, v.K, v.D)
		// fmt.Printf("%.2f %.2f %.2f\n", v.Close, utils.FRound2(k[i]), utils.FRound2(d[i]))
	}
	os.Exit(123)
	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
