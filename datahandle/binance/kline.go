package binance

import (
	"context"

	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/utils"
)

type KlineSrv struct {
	Interval string
}

func NewKlineSrv() *KlineSrv {
	ks := &KlineSrv{
		Interval: cfg.Interval,
	}
	return ks
}

func (this *KlineSrv) Get(limit int) ([]*indicator.Kline, error) {
	bklines, err := client.NewKlinesService().Symbol(cfg.Symbol).
		Interval(cfg.Interval).Limit(limit).Do(context.Background())
	if err != nil {
		return nil, err
	}
	ks := make([]*indicator.Kline, len(bklines)-1)
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
		ks[i] = &kl
	}
	return ks, nil
}

func (this *KlineSrv) WithMa(klines []*indicator.Kline) {
	ma := indicator.NewMa(5)
	ma5 := ma.WithMa(klines)

	ma = indicator.NewMa(10)
	ma10 := ma.WithMa(klines)

	ma = indicator.NewMa(20)
	ma20 := ma.WithMa(klines)

	for i := range klines {
		klines[i].MA5 = ma5[i]
		klines[i].MA10 = ma10[i]
		klines[i].MA20 = ma20[i]
	}
}

func (this *KlineSrv) WithKdj(klines []*indicator.Kline) {
	kdj := indicator.NewKdj(9, 3, 3)
	k, d, _ := kdj.WithKdj(klines)
	for i := range klines {
		klines[i].K = utils.FRound2(k[i])
		klines[i].D = utils.FRound2(d[i])
	}
}
