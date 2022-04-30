package binance

import (
	"context"
	"fmt"

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

func Get(limit int) ([]*indicator.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(cfg.Symbol).
		Interval(cfg.Interval).Limit(limit).Do(context.Background())
	if err != nil {
		return nil, err
	}
	r := []*indicator.Kline{}
	for _, v := range klines[:len(klines)-1] {
		kl := &indicator.Kline{
			OpenTime:  v.OpenTime,
			CloseTime: v.CloseTime,
			Open:      utils.StrToF64(v.Open),
			High:      utils.StrToF64(v.High),
			Low:       utils.StrToF64(v.Low),
			Close:     utils.StrToF64(v.Close),
			Volume:    utils.StrToF64(v.Volume),
			TradeNum:  v.TradeNum,
		}
		r = append(r, kl)
	}
	return r, nil
}

type KlineSrv struct {
	Klines []*indicator.Kline
}

func (k *KlineSrv) Get(limit int) error {
	klines, err := client.NewKlinesService().Symbol(cfg.Symbol).
		Interval(cfg.Interval).Limit(limit).Do(context.Background())
	if err != nil {
		return nil
	}
	for _, v := range klines[:len(klines)-1] {
		kl := &indicator.Kline{
			OpenTime:  v.OpenTime,
			CloseTime: v.CloseTime,
			Open:      utils.StrToF64(v.Open),
			High:      utils.StrToF64(v.High),
			Low:       utils.StrToF64(v.Low),
			Close:     utils.StrToF64(v.Close),
			Volume:    utils.StrToF64(v.Volume),
			TradeNum:  v.TradeNum,
		}
		k.Klines = append(k.Klines, kl)
	}
	return nil
}

type KDKline struct {
	indicator.Kline
	K float64
	D float64
}

type KDKlineSrv struct {
	KlineSrv *KlineSrv
	KDKlines []*KDKline
}

func (k *KDKlineSrv) Get() error {
	for _, v := range k.KlineSrv.Klines {
		fmt.Println(v.Close)
	}
	return nil
}
