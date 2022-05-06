package kdj

import (
	"errors"
	"fmt"
	"time"

	"github.com/LogicHou/bftr/datahandle/binance"
	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

type globalData struct {
	interval    string
	restKlines  []*indicator.Kline
	posAmt      float64
	posQty      float64
	entryPrice  float64
	leverage    float64
	posSide     futures.SideType
	stopLoss    float64
	closeMA     int
	refreshTime map[string]int64
	wsk
}

var _g = &globalData{ //三没用全局变量应付下，@todo 后期通过数据库或者redis持久化
	interval:    "15m",
	restKlines:  nil,
	posAmt:      0,                   // 持仓金额，值等于0的时候表示未持仓
	posQty:      0,                   // 持仓次数，下单后refresh一次数据就+1，todo 后期如果使用多个goroutine监控则需要在+1的时候先上锁
	entryPrice:  0,                   // 开仓均价
	leverage:    0,                   // 当前杠杆倍数
	posSide:     futures.SideTypeBuy, // 持仓的买卖方向
	stopLoss:    0,                   // 止损数值
	closeMA:     20,                  // 止损均线值
	refreshTime: map[string]int64{"30m": 3603000, "15m": 1803000, "5m": 603000},
}

type wsk struct {
	h   float64
	l   float64
	c   float64
	v   float64
	E   int64
	cma float64
}

type KDJ struct {
}

func NewStrategyKDJ() *KDJ {
	return &KDJ{}
}

func (this *KDJ) Run() error {
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
	tradeSrv := binance.NewTradeSrv()
	go func() {
		lastRsk := _g.restKlines[len(_g.restKlines)-1]

		for k := range bclient.WsKlineCh {
			_g.wsk.c = utils.StrToF64(k.Kline.Close)
			_g.wsk.h = utils.StrToF64(k.Kline.High)
			_g.wsk.l = utils.StrToF64(k.Kline.Low)
			// _g.wsk.v = utils.StrToF64(k.Kline.Volume) // 用不到成交量先注释掉
			_g.wsk.E = k.Time
			ma := indicator.NewMa(_g.closeMA)
			_g.wsk.cma = ma.CurrentMa(_g.restKlines, _g.wsk.c)

			if (_g.wsk.E - lastRsk.OpenTime) > _g.refreshTime[_g.interval] {
				refreshSomeData()
				if _g.restKlines[len(_g.restKlines)-1].OpenTime == lastRsk.OpenTime {
					time.Sleep(6 * time.Second)
					continue
				}
				lastRsk = _g.restKlines[len(_g.restKlines)-1]
				_g.posQty += 1
			}
			fmt.Println(_g.wsk.c)
			// 开仓逻辑
			// @todo 判断下单条件是否满足
			// @todo 判断交易方向
			// @todo 调用下单（市价单）方法
			// @todo 判断止损条件是否满足
			// @todo 调用止损方法
			_g.stopLoss, _ = findFrontHigh(_g.restKlines, futures.SideTypeSell) // 开仓的时候才计算这个止损

			// 止盈逻辑，@todo 迁移到独立的goroutine中
			if _g.posQty >= 3 {
				switch _g.posSide {
				case futures.SideTypeBuy:
					if _g.wsk.c < _g.wsk.cma {
						tradeSrv.ClosePosition(_g.posAmt)
						// @todo 记录日志，重置一些数据
					}
				case futures.SideTypeSell:
					if _g.wsk.c > _g.wsk.cma {
						tradeSrv.ClosePosition(_g.posAmt)
						// @todo 记录日志，重置一些数据
					}
				}
			}

			// 止损逻辑，@todo 迁移到独立的goroutine中
			switch _g.posSide {
			case futures.SideTypeBuy:
				if _g.wsk.c < _g.stopLoss {
					tradeSrv.ClosePosition(_g.posAmt)
					// @todo 记录日志，重置一些数据
				}
			case futures.SideTypeSell:
				if _g.wsk.c > _g.stopLoss {
					tradeSrv.ClosePosition(_g.posAmt)
					// @todo 记录日志，重置一些数据
				}
			}
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
	_g.interval = klineSrv.Interval
	_g.restKlines, err = klineSrv.Get(41)
	if err != nil {
		return err
	}
	klineSrv.WithKdj(_g.restKlines)
	klineSrv.WithMa(_g.restKlines)

	return nil
}

func findFrontHigh(klines []*indicator.Kline, posSide futures.SideType) (float64, error) {
	if posSide == futures.SideTypeBuy {
		ksLen := len(klines)
		low := klines[ksLen-1].Low
		for i := range klines {
			if klines[ksLen-i-1].Low < low &&
				klines[ksLen-i-2].Low > klines[ksLen-i-1].Low {
				return klines[ksLen-i-1].Low, nil
			}
		}
		return 0.0, errors.New("not found stoploss condition")
	}

	if posSide == futures.SideTypeSell {
		ksLen := len(klines)
		high := klines[ksLen-1].High
		for i := range klines {
			if klines[ksLen-i-1].High > high &&
				klines[ksLen-i-2].High < klines[ksLen-i-1].High {
				return klines[ksLen-i-1].High, nil
			}
		}
		return 0.0, errors.New("not found stoploss condition")
	}

	return 0.0, errors.New("not found stoploss condition")
}
