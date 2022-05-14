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
	//refresh some data
	err = ts.updateHandler()
	if err != nil {
		errChan <- err
	}
	td := ts.getHandler()
	// fmt.Println(td.HistKlines[39])

	log.Println("Info: Data initialization succeeded!")

	go func() {
		log.Println("listening transactions...")
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
				ts.updateHandler()
				if td.HistKlines[len(td.HistKlines)-1].OpenTime == lastRsk.OpenTime {
					time.Sleep(6 * time.Second) // @todo 改成goroutine形式
					continue
				}
				lastRsk = td.HistKlines[len(td.HistKlines)-1]
				log.Printf("PosAmt:%f, EntryPrice:%f, Leverage:%f, PosSide:%s StopLoss:td.StopLoss:%f\n", td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide, td.StopLoss)
			}

			// 开仓逻辑
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
				curK, _, _ := kdj.WithKdj(tmpks)

				// 开仓点
				if ts.openCondition(td.PosSide, curK[len(curK)-1], lastRsk) {
					log.Println("beging creating order...")
					qty := ts.tradeSrv.CalcMqrginQty(td.Wsk.C)

					switch td.PosSide {
					case futures.SideTypeBuy:
						td.StopLoss, err = ts.findFrontLow(td.HistKlines, futures.SideTypeBuy)
						err = ts.tradeSrv.CreateMarketOrder(futures.SideTypeBuy, qty, td.StopLoss-1)
					case futures.SideTypeSell:
						td.StopLoss, err = ts.findFrontLow(td.HistKlines, futures.SideTypeSell)
						err = ts.tradeSrv.CreateMarketOrder(futures.SideTypeSell, qty, td.StopLoss+1)
					}
					if err != nil {
						errChan <- err
					}
					td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide, err = ts.tradeSrv.GetPostionRisk()
					if err != nil {
						errChan <- err
					}
				}
				continue
			}

			// 止盈逻辑
			if td.PosQty > ts.tradeSrv.PosQtyUlimit {
				log.Println("beging take profit")
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
			log.Println("beging stop loss")
			switch td.PosSide {
			case futures.SideTypeBuy:
				if td.Wsk.C < td.StopLoss {
					ts.tradeSrv.ClosePosition(td.PosAmt)
					ts.resetHandler()
					// @todo 记录日志，重置一些数据
				}
			case futures.SideTypeSell:
				if td.Wsk.C > td.StopLoss {
					ts.tradeSrv.ClosePosition(td.PosAmt)
					ts.resetHandler()
					// @todo 记录日志，重置一些数据
				}
			}

			td.PosQty += 1
		}
	}()
	select {
	case err = <-errChan:
		return nil, err
	case <-time.After(time.Second):
		return errChan, nil
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

	return nil
}

func (ts *TradeServer) resetHandler() error {
	ts.s.Reset()
	return nil
}

func (ts *TradeServer) openCondition(side futures.SideType, curK float64, lastRsk *indicator.Kline) bool {
	switch side {
	case futures.SideTypeBuy:
		if lastRsk.K < ts.tradeSrv.OpenK1 && curK > ts.tradeSrv.OpenK1 {
			return true
		}
		if lastRsk.K < ts.tradeSrv.OpenK3 && curK > ts.tradeSrv.OpenK3 {
			return true
		}
	case futures.SideTypeSell:
		if lastRsk.K > ts.tradeSrv.OpenK2 && curK < ts.tradeSrv.OpenK2 {
			return true
		}
		if lastRsk.K > ts.tradeSrv.OpenK3 && curK < ts.tradeSrv.OpenK3 {
			return true
		}
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
