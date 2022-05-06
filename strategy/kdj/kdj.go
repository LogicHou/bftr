package strategy

import (
	"fmt"
	"time"

	"github.com/LogicHou/bftr/datahandle/binance"
	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

var (
	_interval    string             = "15m"
	_restKlines  []*indicator.Kline = nil
	_posAmt      float64            = 0  // 持仓金额，值等于0的时候表示未持仓
	_entryPrice  float64            = 0  // 开仓均价
	_leverage    float64            = 0  // 当前杠杆倍数
	_posSide     futures.SideType        // 持仓的买卖方向
	_stopLoss    float64            = 0  // 止损数值
	_closeMA     int                = 20 // 止损均线值
	_refreshTime                    = map[string]int64{"30m": 3603000, "15m": 1803000, "5m": 603000}
	_wsk                            = struct {
		h  float64
		l  float64
		c  float64
		v  float64
		E  int64
		st float64
	}{0, 0, 0, 0, 0, 0}
)

type KDJ struct {
}

func NewStrategyKDJ() *KDJ {
	return &KDJ{}
}

func (this *KDJ) Run() error {
	fmt.Println("strategyKDJ run")
	err := refreshSomeData()
	if err != nil {
		return err
	}
	fmt.Println("Info: Data initialization succeeded!")

	// for _, v := range _restKlines {
	// 	fmt.Println("bbb", v.Close, v.K, v.MA5, v.MA20)
	// }
	bclient, err := binance.NewClient()
	if err != nil {
		err = fmt.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		return err
	}
	go func() {
		lastRsk := _restKlines[len(_restKlines)-1]
		for k := range bclient.WsKlineCh {
			_wsk.c = utils.StrToF64(k.Kline.Close)
			_wsk.h = utils.StrToF64(k.Kline.High)
			_wsk.l = utils.StrToF64(k.Kline.Low)
			// _wsk.v = utils.StrToF64(k.Kline.Volume) // 用不到成交量先注释掉
			_wsk.E = k.Time
			ma := indicator.NewMa(_closeMA)
			_wsk.st = ma.CurrentMa(_restKlines, _wsk.c)
			fmt.Println(_wsk.st)

			if (_wsk.E - lastRsk.OpenTime) > _refreshTime[_interval] {
				refreshSomeData()
			}
			// @todo 判断下单条件是否满足
			// @todo 判断交易方向
			// @todo 调用下单（市价单）方法
			// @todo 判断止损条件是否满足
			// @todo 调用止损方法
		}
	}()

	for {
		<-bclient.DoneC
		return nil
	}
}

func refreshSomeData() error {
	var err error

	klineSrv := binance.NewKlineSrv()
	_interval = klineSrv.Interval
	_restKlines, err = klineSrv.Get(41)
	if err != nil {
		return err
	}
	klineSrv.WithKdj(_restKlines)
	klineSrv.WithMa(_restKlines)

	return nil
}
