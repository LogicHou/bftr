package binance

import (
	"context"
	"errors"

	"github.com/LogicHou/bftr/datahandle"
	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/utils"
)

type KlineSrv struct {
}

func NewKlineSrv() datahandle.Kline {
	return &KlineSrv{}
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

func (this *KlineSrv) WithKdj(klines []*indicator.Kline) error {
	kdj := indicator.NewKdj(9, 3, 3)
	k, d, _ := kdj.WithKdj(klines)
	if len(klines) != len(k) {
		return errors.New("the comparison objects are not equal")
	}
	for i := range klines {
		klines[i].K = utils.FRound2(k[i])
		klines[i].D = utils.FRound2(d[i])
	}
	return nil
}
