package server

import (
	"errors"
	"log"
	"time"

	bds "github.com/LogicHou/bftr/datasrv/binance"
	"github.com/LogicHou/bftr/indicator"
	"github.com/LogicHou/bftr/store"
	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

type TradeServer struct {
	s        store.Store
	srv      *bds.Server
	tradeSrv *bds.TradeSrv
}

func NewTradeServer(s store.Store) *TradeServer {
	srv := &TradeServer{
		s: s,
		srv: &bds.Server{
			WskChan: make(chan *futures.WsKlineEvent),
		},
		tradeSrv: bds.NewTradeSrv(),
	}

	return srv
}

func (ts *TradeServer) ListenAndServe() (<-chan error, error) {
	var err error
	errChan := make(chan error)
	go func() {
		err = ts.srv.Serve()
		errChan <- err
	}()

	select {
	case err = <-errChan:
		return nil, err
	case <-time.After(time.Second):
		return errChan, nil
	}
}

func (ts *TradeServer) ListenAndMonitor() (<-chan error, error) {
	var err error
	errChan := make(chan error)

	go func() {
		//refresh some data
		err = ts.updateHandler()
		if err != nil {
			errChan <- err
		}
		td := ts.getHandler()

		log.Println("Info: Data initialization succeeded!")

		log.Println("listening transactions...")
		// tradeLock := false
		lastRsk := td.HistKlines[len(td.HistKlines)-1]
		cmai := indicator.NewMa(ts.tradeSrv.CloseMa)
		omai := indicator.NewMa(ts.tradeSrv.OpenSideMa)
		kdj := indicator.NewKdj(9, 3, 3)

		for k := range ts.srv.WskChan {
			td.Wsk.C = utils.StrToF64(k.Kline.Close)
			td.Wsk.H = utils.StrToF64(k.Kline.High)
			td.Wsk.L = utils.StrToF64(k.Kline.Low)
			td.Wsk.E = k.Time
			td.Wsk.Cma, err = cmai.CurMa(td.HistKlines, td.Wsk.C)
			if err != nil {
				errChan <- err
			}

			// 刷新数据节点
			if (td.Wsk.E - lastRsk.OpenTime) > td.RefreshTime[ts.tradeSrv.Interval] {
				err = ts.updateHandler()
				if err != nil {
					errChan <- err
				}
				if td.HistKlines[len(td.HistKlines)-1].OpenTime == lastRsk.OpenTime {
					time.Sleep(6 * time.Second) // @todo 改成goroutine形式
					continue
				}

				lastRsk = td.HistKlines[len(td.HistKlines)-1]
				if td.PosAmt != 0 {
					td.PosQty += 1
				}
				// tradeLock = false
			}

			// 开仓逻辑
			// if td.PosAmt == 0 && tradeLock == false {
			if td.PosAmt == 0 {
				oma, err := omai.CurMa(td.HistKlines, td.Wsk.C)
				if err != nil {
					errChan <- err
				}
				if td.Wsk.C > oma {
					td.PosSide = futures.SideTypeBuy
				} else if td.Wsk.C < oma {
					td.PosSide = futures.SideTypeSell
				} else {
					continue
				}

				// @todo 这里不是很好的解决方案，后面再改进
				tmpks := append(td.HistKlines, &indicator.Kline{
					Close: td.Wsk.C,
					High:  td.Wsk.H,
					Low:   td.Wsk.L,
				})
				curKs, _, _ := kdj.WithKdj(tmpks)
				curK := curKs[len(curKs)-1]

				// 开仓点
				if ts.openCondition(td.PosSide, curK, lastRsk) {
					log.Println("beging creating order...")
					qty, err := ts.tradeSrv.CalcMqrginQty(td.Wsk.C)
					if err != nil {
						errChan <- err
					}

					switch td.PosSide {
					case futures.SideTypeBuy:
						td.StopLoss, err = ts.findFrontLow(td.HistKlines, futures.SideTypeBuy)
						err = ts.tradeSrv.CreateMarketOrder(futures.SideTypeBuy, qty, td.StopLoss)
					case futures.SideTypeSell:
						td.StopLoss, err = ts.findFrontLow(td.HistKlines, futures.SideTypeSell)
						err = ts.tradeSrv.CreateMarketOrder(futures.SideTypeSell, qty, td.StopLoss)
					}
					if err != nil {
						errChan <- err
					}
					td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide, err = ts.tradeSrv.GetPostionRisk()

					td.Openk = curK
					log.Printf("CreateMarketOrder - PosSide:%s PosAmt:%f Openk:%f EntryPrice:%f StopLoss:%f\n", td.PosSide, td.PosAmt, td.Openk, td.EntryPrice, td.StopLoss)

					if err != nil {
						errChan <- err
					}
					// tradeLock = true
				}
				continue
			}

			// 止盈逻辑
			if td.PosQty > ts.tradeSrv.PosQtyUlimit {
				switch td.PosSide {
				case futures.SideTypeBuy:
					if td.Wsk.C < td.Wsk.Cma {
						err = ts.closePosition()
					}
				case futures.SideTypeSell:
					if td.Wsk.C > td.Wsk.Cma {
						err = ts.closePosition()
					}
				}
				if err != nil {
					errChan <- err
				}
				continue
			}

			// 止损逻辑
			if ts.closeCondition(lastRsk) {
				ts.closePosition()
			}

		}
	}()

	return errChan, nil
}

func (ts *TradeServer) SafetyQuit() {
	log.Println("trigger SafetyQuit")
	td := ts.s.Get()
	if td.PosAmt != 0 {
		ts.closePosition()
	}
}

func (ts *TradeServer) getHandler() *store.Trader {
	return ts.s.Get()
}

func (ts *TradeServer) updateHandler() error {
	klineSrv := bds.NewKlineSrv()
	histKlines, err := klineSrv.Get()
	if err != nil {
		return err
	}
	klineSrv.WithKdj(histKlines)
	klineSrv.WithMa(histKlines)
	ts.s.Update(histKlines)

	td := ts.s.Get()
	td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide, err = ts.tradeSrv.GetPostionRisk()
	if err != nil {
		return err
	}
	if td.PosAmt != 0 && td.StopLoss == 0 {
		stopPrice, orderTime, err := ts.tradeSrv.GetOpenOrder()
		td.PosQty = int((time.Now().UnixMilli() - orderTime) / (td.RefreshTime[ts.tradeSrv.Interval] / 2))
		if err != nil {
			return err
		}
		if td.PosAmt > 0 {
			td.PosSide = futures.SideTypeBuy
			td.StopLoss = stopPrice
		}
		if td.PosAmt < 0 {
			td.PosSide = futures.SideTypeBuy
			td.StopLoss = stopPrice
		}
	}

	log.Printf("Refreshed: PosSide:%s PosAmt:%f PosQty:%d EntryPrice:%f Leverage:%f StopLoss:%f\n", td.PosSide, td.PosAmt, td.PosQty, td.EntryPrice, td.Leverage, td.StopLoss)

	return nil
}

func (ts *TradeServer) resetHandler() error {
	ts.s.Reset()
	return nil
}

func (ts *TradeServer) openCondition(side futures.SideType, curK float64, lastRsk *indicator.Kline) bool {
	offset := 1.00
	switch side {
	case futures.SideTypeBuy:
		if lastRsk.K < ts.tradeSrv.OpenK1 && curK > (ts.tradeSrv.OpenK1+offset) {
			return true
		}
		if lastRsk.K < ts.tradeSrv.OpenK3 && curK > (ts.tradeSrv.OpenK3+offset) {
			return true
		}
	case futures.SideTypeSell:
		if lastRsk.K > ts.tradeSrv.OpenK2 && curK < (ts.tradeSrv.OpenK2-offset) {
			return true
		}
		if lastRsk.K > ts.tradeSrv.OpenK3 && curK < (ts.tradeSrv.OpenK3-offset) {
			return true
		}
	}
	return false
}

func (ts *TradeServer) closeCondition(lastRsk *indicator.Kline) bool {
	td := ts.s.Get()
	switch td.PosSide {
	case futures.SideTypeBuy:
		if td.Wsk.C < td.StopLoss {
			return true
		}
		// if td.PosQty == 2 && lastRsk.K < td.Openk {
		// 	return true
		// }
	case futures.SideTypeSell:
		if td.Wsk.C > td.StopLoss {
			return true
		}
		// if td.PosQty == 2 && lastRsk.K > td.Openk {
		// 	return true
		// }
	}
	return false
}

func (ts *TradeServer) findFrontLow(klines []*indicator.Kline, posSide futures.SideType) (float64, error) {
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

func (ts *TradeServer) closePosition() error {
	td := ts.s.Get()
	err := ts.tradeSrv.ClosePosition(td.PosAmt)
	if err != nil {
		return err
	}

	td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide, err = ts.tradeSrv.GetPostionRisk()
	if err != nil {
		return err
	}

	// reset some datas
	ts.s.Reset()

	return nil
}
