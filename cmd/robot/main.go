package main

import (
	"fmt"
	"time"

	"github.com/LogicHou/bftr/datasource/binance"
	_ "github.com/LogicHou/bftr/internal/store"
)

func main() {
	KlineSrv, err := binance.New()
	if err != nil {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
	}
	err = KlineSrv.Get(145)
	if err != nil {
		fmt.Println(err)
	}
	KlineSrv.WithKdj()
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range KlineSrv.Klines {
		fmt.Printf("%.2f %.2f %.2f\n", v.Close, v.K, v.D)
		// fmt.Printf("%.2f %.2f %.2f\n", v.Close, utils.FRound2(k[i]), utils.FRound2(d[i]))
	}

	// var klineCh = make(chan *futures.WsKlineEvent)

	go func() {
		for k := range KlineSrv.KlineCh {
			fmt.Println(k.Kline.Close)
		}
	}()

	for {
		<-KlineSrv.DoneC
	}

}
