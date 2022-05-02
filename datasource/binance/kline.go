package binance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/internal/config"
	"github.com/LogicHou/bftr/utils"
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

type KlineSrv struct {
	Client  *futures.Client
	Klines  []*indicator.Kline
	KlineCh chan *futures.WsKlineEvent
	DoneC   chan struct{}
	StopC   chan struct{}
}

func New() (*KlineSrv, error) {
	ks := &KlineSrv{
		Client:  client,
		KlineCh: make(chan *futures.WsKlineEvent),
	}
	doneC, stopC, err := futures.WsKlineServe(cfg.Binance.Symbol, cfg.Binance.Interval,
		func(event *futures.WsKlineEvent) {
			ks.KlineCh <- event
		},
		func(err error) {
			log.Println(time.Now().Format("2006-01-02 15:04:05"), "errmsg:", err)
		})
	if err != nil {
		err = fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		return &KlineSrv{}, err
	}
	ks.DoneC = doneC
	ks.StopC = stopC
	return ks, nil
}

func (this *KlineSrv) Get(limit int) error {
	bklines, err := this.Client.NewKlinesService().Symbol(cfg.Symbol).
		Interval(cfg.Interval).Limit(limit).Do(context.Background())
	if err != nil {
		return err
	}
	this.Klines = make([]*indicator.Kline, len(bklines)-1)
	for i, v := range bklines[:len(bklines)-1] {
		kl := indicator.Kline{
			OpenTime:  v.OpenTime,
			CloseTime: v.CloseTime,
			Open:      utils.StrToF64(v.Open),
			High:      utils.StrToF64(v.High),
			Low:       utils.StrToF64(v.Low),
			Close:     utils.StrToF64(v.Close),
			Volume:    utils.StrToF64(v.Volume),
			TradeNum:  v.TradeNum,
		}
		this.Klines[i] = &kl
	}
	return nil
}

func (this *KlineSrv) WithKdj() error {
	kdj := indicator.NewKdj(9, 3, 3)
	k, d, _ := kdj.WithKdj(this.Klines)
	if len(this.Klines) != len(k) {
		return errors.New("the comparison objects are not equal")
	}
	for i := range this.Klines {
		this.Klines[i].K = utils.FRound2(k[i])
		this.Klines[i].D = utils.FRound2(d[i])
	}
	return nil
}
