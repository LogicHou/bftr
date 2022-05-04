package strategy

import (
	"fmt"
	"time"

	"github.com/LogicHou/bftr/datahandle"
	"github.com/LogicHou/bftr/datahandle/binance"
)

type KDJ struct {
}

func NewStrategyKDJ() *KDJ {
	return &KDJ{}
}

var kline *datahandle.Kline
var trade *datahandle.Trade

func (this *KDJ) Run() error {
	fmt.Println("strategyKDJ run")
	bclient, err := binance.NewClient()
	if err != nil {
		err = fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		return err
	}
	go func() {
		for k := range bclient.WsKlineCh {
			// @todo 刷新历史K线数据
			// @todo 判断下单条件是否满足
			// @todo 判断交易方向
			// @todo 调用下单（市价单）方法
			// @todo 判断止损条件是否满足
			// @todo 调用止损方法
			fmt.Println(k.Event, k.Kline.Close)
		}
	}()

	for {
		<-bclient.DoneC
		return nil
	}
}
