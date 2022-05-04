package binance

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LogicHou/bftr/datahandle"
	"github.com/LogicHou/bftr/internal/config"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

var (
	cfg    *config.Cfg
	client *futures.Client
)

func init() {
	cfg = config.Get()
	client = binance.NewFuturesClient(cfg.Binance.ApiKey, cfg.Binance.SecretKey)
	client.NewSetServerTimeService().Do(context.Background())
}

type binanceClient struct{}

func NewClient() datahandle.Client {
	return &binanceClient{}
}

func (b *binanceClient) Start() error {
	// KlineSrv, err := NewKlineSrv()
	// if err != nil {
	// 	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
	// }
	// err = KlineSrv.Get(145)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// KlineSrv.WithKdj()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for _, v := range KlineSrv.Klines {
	// 	fmt.Printf("%.2f %.2f %.2f\n", v.Close, v.K, v.D)
	// 	// fmt.Printf("%.2f %.2f %.2f\n", v.Close, utils.FRound2(k[i]), utils.FRound2(d[i]))
	// }
	var klineCh = make(chan *futures.WsKlineEvent)
	doneC, stopC, err := futures.WsKlineServe(cfg.Binance.Symbol, cfg.Binance.Interval,
		func(event *futures.WsKlineEvent) {
			klineCh <- event
		},
		func(err error) {
			log.Println(time.Now().Format("2006-01-02 15:04:05"), "errmsg:", err)
		})
	if err != nil {
		err = fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		return err
	}

	go func() {
		for k := range klineCh {
			fmt.Println(k.Kline.Close)
		}
	}()

	for {
		<-doneC
		<-stopC
	}
}
