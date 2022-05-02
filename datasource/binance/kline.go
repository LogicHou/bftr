package binance

import (
	"context"
	"errors"

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
	client = binance.NewFuturesClient(cfg.ApiKey, cfg.SecretKey)
	client.NewSetServerTimeService().Do(context.Background())
}

type KlineSrv struct {
	Klines []indicator.Kline
}

func (this *KlineSrv) Get(limit int) error {
	bklines, err := client.NewKlinesService().Symbol(cfg.Symbol).
		Interval(cfg.Interval).Limit(limit).Do(context.Background())
	if err != nil {
		return err
	}
	this.Klines = make([]indicator.Kline, len(bklines)-1)
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
		this.Klines[i] = kl
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
