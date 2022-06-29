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

		log.Printf("Data initialization succeeded: PosSide:%s PosAmt:%f PosQty:%d EntryPrice:%f Leverage:%f StopLoss:%f\n", td.PosSide, td.PosAmt, td.PosQty, td.EntryPrice, td.Leverage, td.StopLoss)

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
			if (td.Wsk.E - lastRsk.CloseTime) > td.RefreshTime[ts.tradeSrv.Interval] {
				err = ts.updateHandler()
				if err != nil {
					errChan <- err
				}
				if td.HistKlines[len(td.HistKlines)-1].OpenTime == lastRsk.OpenTime {
					log.Printf("updateHandler delay")
					time.Sleep(6 * time.Second) // TODO 改成goroutine形式
					continue
				}

				lastRsk = td.HistKlines[len(td.HistKlines)-1]
				if td.PosAmt != 0 {
					td.PosQty += 1
				}
				log.Printf("Refreshed: PosSide:%s PosAmt:%f PosQty:%d EntryPrice:%f Leverage:%f StopLoss:%f\n", td.PosSide, td.PosAmt, td.PosQty, td.EntryPrice, td.Leverage, td.StopLoss)

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

				// TODO 这里不是很好的解决方案，后面再改进
				tmpks := append(td.HistKlines, &indicator.Kline{
					Close: td.Wsk.C,
					High:  td.Wsk.H,
					Low:   td.Wsk.L,
				})
				curKs, _, _ := kdj.WithKdj(tmpks)
				curK := curKs[len(curKs)-1]

				// 开仓点
				if ts.openCondition(td.PosSide, curK, td.HistKlines) {
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

					if td.PosAmt == 0 {
						log.Println("GetPostionRisk may be failed, the data:", td.PosAmt, td.EntryPrice, td.Leverage, td.PosSide)
						td.PosAmt = qty
					}

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
		td.PosQty = int((time.Now().UnixMilli()-orderTime)/(td.RefreshTime[ts.tradeSrv.Interval]/2)) + 1
		if err != nil {
			return err
		}
		if td.PosAmt > 0 {
			td.PosSide = futures.SideTypeBuy
			td.StopLoss = stopPrice
		}
		if td.PosAmt < 0 {
			td.PosSide = futures.SideTypeSell
			td.StopLoss = stopPrice
		}
	}

	return nil
}

func (ts *TradeServer) resetHandler() error {
	ts.s.Reset()
	return nil
}

func (ts *TradeServer) openCondition(side futures.SideType, curK float64, kls []*indicator.Kline) bool {
	kOffset := ts.tradeSrv.KOffset
	crossOffset := ts.tradeSrv.CrossOffset
	preKl1 := kls[len(kls)-1]
	preKl2 := kls[len(kls)-2]
	goldCross, deadCross := false, false

	switch side {
	case futures.SideTypeBuy:
		if preKl1.K < ts.tradeSrv.OpenK1 && curK > (ts.tradeSrv.OpenK1+kOffset) {
			return true
		}
		if preKl1.K < ts.tradeSrv.OpenK3 && curK > (ts.tradeSrv.OpenK3+kOffset) {
			for i := len(kls) - 1; i > 0; i-- {
				if kls[i].K > ts.tradeSrv.OpenK3 && kls[i-1].K < ts.tradeSrv.OpenK3 {
					break
				}
				if kls[i].K > ts.tradeSrv.OpenK3 && (kls[i].K-crossOffset > kls[i].D && kls[i-1].K < kls[i-1].D) {
					goldCross = true
				}
				if kls[i].K > ts.tradeSrv.OpenK3 && (kls[i].K < kls[i].D && kls[i-1].K > kls[i-1].D) {
					deadCross = true
				}
			}
			return !(goldCross && deadCross)
		}
		if preKl2.K < ts.tradeSrv.OpenK1 && preKl1.K > ts.tradeSrv.OpenK1 {
			return true
		}
	case futures.SideTypeSell:
		if preKl1.K > ts.tradeSrv.OpenK2 && curK < (ts.tradeSrv.OpenK2-kOffset) {
			return true
		}
		if preKl1.K > ts.tradeSrv.OpenK3 && curK < (ts.tradeSrv.OpenK3-kOffset) {
			for i := len(kls) - 1; i > 0; i-- {
				if kls[i].K < ts.tradeSrv.OpenK3 && kls[i-1].K > ts.tradeSrv.OpenK3 {
					break
				}
				if kls[i].K < ts.tradeSrv.OpenK3 && (kls[i].K+crossOffset < kls[i].D && kls[i-1].K > kls[i-1].D) {
					goldCross = true
				}
				if kls[i].K < ts.tradeSrv.OpenK3 && (kls[i].K > kls[i].D && kls[i-1].K < kls[i-1].D) {
					deadCross = true
				}
			}
			return !(goldCross && deadCross)
		}
		if preKl2.K > ts.tradeSrv.OpenK2 && preKl1.K < ts.tradeSrv.OpenK2 {
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

func (ts *TradeServer) findFrontLow(kls []*indicator.Kline, posSide futures.SideType) (float64, error) {
	td := ts.s.Get()
	if posSide == futures.SideTypeBuy {
		low := td.Wsk.L
		for i := len(kls) - 1; i > 0; i-- {
			if kls[i].Low < low && kls[i].Low < kls[i-1].Low {
				return kls[i].Low, nil
			}
		}
		return 0.0, errors.New("not found stoploss condition")
	}

	if posSide == futures.SideTypeSell {
		high := td.Wsk.H
		for i := len(kls) - 1; i > 0; i-- {
			if kls[i].High > high && kls[i].High > kls[i-1].High {
				return kls[i].High, nil
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
