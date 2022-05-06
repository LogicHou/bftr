package binance

import (
	"context"
	"fmt"
	"time"

	"github.com/LogicHou/bftr/internal/config"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
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

type binanceClient struct {
	WsKlineCh chan *futures.WsKlineEvent
	DoneC     chan struct{}
	StopC     chan struct{}
}

func NewClient() (*binanceClient, error) {
	bc := &binanceClient{
		WsKlineCh: make(chan *futures.WsKlineEvent),
	}
	doneC, stopC, err := futures.WsKlineServe(cfg.Binance.Symbol, cfg.Binance.Interval,
		func(event *futures.WsKlineEvent) {
			bc.WsKlineCh <- event
		},
		func(err error) {
			err = errors.Errorf(time.Now().Format("2006-01-02 15:04:05"), "errmsg:", err)
		})
	if err != nil {
		err = fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		return nil, err
	}
	bc.DoneC = doneC
	bc.StopC = stopC
	return bc, nil
}
