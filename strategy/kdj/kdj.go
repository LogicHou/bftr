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
	margin      float64
	interval    string
	leverage    float64
	restKlines  []*indicator.Kline
	posAmt      float64
	posQty      float64
	entryPrice  float64
	posSide     futures.SideType
	stopLoss    float64
	refreshTime map[string]int64
	wsk
}

type wsk struct {
	h   float64
	l   float64
	c   float64
	v   float64
	E   int64
	cma float64
}

var _g = &globalData{ //三没用全局变量应付下，@todo 后期通过数据库或者redis持久化
	interval:    "15m",
	restKlines:  nil,
	posAmt:      0,                   // 持仓金额，值等于0的时候表示未持仓
	posQty:      0,                   // 持仓次数，下单后refresh一次数据就+1，todo 后期如果使用多个goroutine监控则需要在+1的时候先上锁
	entryPrice:  0,                   // 开仓均价
	leverage:    0,                   // 当前杠杆倍数
	posSide:     futures.SideTypeBuy, // 持仓的买卖方向，默认为买方向
	stopLoss:    0,                   // 止损数值
	refreshTime: map[string]int64{"30m": 3603000, "15m": 1803000, "5m": 603000},
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

	lastRsk := _g.restKlines[len(_g.restKlines)-1]

	for k := range bclient.WsKlineCh {
		_g.wsk.c = utils.StrToF64(k.Kline.Close)
		_g.wsk.h = utils.StrToF64(k.Kline.High)
		_g.wsk.l = utils.StrToF64(k.Kline.Low)
		// _g.wsk.v = utils.StrToF64(k.Kline.Volume) // 用不到成交量先注释掉
		_g.wsk.E = k.Time
		ma := indicator.NewMa(tradeSrv.CloseMa)
		_g.wsk.cma = ma.CurMa(_g.restKlines, _g.wsk.c)

		// 刷新数据节点
		if (_g.wsk.E - lastRsk.OpenTime) > _g.refreshTime[_g.interval] {
			refreshSomeData()
			if _g.restKlines[len(_g.restKlines)-1].OpenTime == lastRsk.OpenTime {
				time.Sleep(6 * time.Second) // @godo 改成goroutine形式
				continue
			}
			lastRsk = _g.restKlines[len(_g.restKlines)-1]
			_g.posQty += 1
		}

		// 开仓逻辑
		// @todo 现在的是过程方式写法，行情数据会在判断流中被阻塞，后面改成通过goroutine联动
		if _g.posAmt == 0 {
			ma := indicator.NewMa(tradeSrv.OpenSideMa)
			oma := ma.CurMa(_g.restKlines, _g.wsk.c)
			if _g.wsk.c > oma {
				_g.posSide = futures.SideTypeBuy
			} else if _g.wsk.c < oma {
				_g.posSide = futures.SideTypeSell
			} else {
				continue
			}

			kdj := indicator.NewKdj(9, 3, 3)
			// @todo 这里不是很好的解决方案，后面再改进
			tmpks := append(_g.restKlines, &indicator.Kline{
				Close: _g.wsk.c,
				High:  _g.wsk.h,
				Low:   _g.wsk.l,
			})
			curK, _, _ := kdj.WithKdj(tmpks)

			// 开仓点
			if openCondition(_g.posSide, curK[len(curK)-1], lastRsk, tradeSrv) {
				_g.stopLoss, err = findFrontHigh(_g.restKlines, futures.SideTypeSell)
				if err != nil {
					return err
				}

				qty := tradeSrv.CalcMqrginQty(_g.margin, _g.leverage, _g.wsk.c)
				switch _g.posSide {
				case futures.SideTypeBuy:
					tradeSrv.CreateMarketOrder(futures.SideTypeBuy, qty, _g.stopLoss)
				case futures.SideTypeSell:
					tradeSrv.CreateMarketOrder(futures.SideTypeSell, qty, _g.stopLoss)
				}
			}
			continue
		}

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

	for {
		<-bclient.DoneC
		return nil
	}
}

func openCondition(side futures.SideType, curK float64, lastRsk *indicator.Kline, tradeSrv *binance.TradeSrv) bool {
	// @todo1 补充符合openK3的开仓条件
	switch side {
	case futures.SideTypeBuy:
		if lastRsk.K < tradeSrv.OpenK1 && curK > tradeSrv.OpenK1 {
			return true
		}
	case futures.SideTypeSell:
		if lastRsk.K > tradeSrv.OpenK2 && curK < tradeSrv.OpenK2 {
			return true
		}
	}
	return false
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
