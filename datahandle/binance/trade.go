package binance

import (
	"context"
	"log"
	"math"

	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

type TradeSrv struct {
	Margin     float64
	OpenSideMa int
	CloseMa    int
	OpenK1     float64
	OpenK2     float64
	OpenK3     float64
}

func NewTradeSrv() *TradeSrv {
	return &TradeSrv{
		OpenSideMa: cfg.OpenSideMa,
		CloseMa:    cfg.CloseMa,
		OpenK1:     cfg.OpenK1,
		OpenK2:     cfg.OpenK2,
		OpenK3:     cfg.OpenK3,
	}
}

func (b *TradeSrv) CalcMqrginQty(margin float64, leverage float64, curClose float64) float64 {
	if margin > 0 {
		return utils.FRound(margin * leverage / curClose)
	}

	return 0
}

func (b *TradeSrv) CreateMarketOrder(sideType futures.SideType, qty float64, firstStopLoss float64) {
	// 取消所有挂单
	err := client.NewCancelAllOpenOrdersService().Symbol(cfg.Symbol).Do(context.Background())
	if err != nil {
		log.Println("createMarketOrder - 1:", err)
		return
	}
	var sideStop futures.SideType
	switch sideType {
	case futures.SideTypeBuy:
		sideStop = futures.SideTypeSell
	case futures.SideTypeSell:
		sideStop = futures.SideTypeBuy
	}

	// 预埋止损单 RestAPI
	order, err := client.NewCreateOrderService().Symbol(cfg.Symbol).
		Side(sideStop).Type("STOP_MARKET").
		ClosePosition(true).StopPrice(utils.F64ToStr(utils.FRound2(firstStopLoss))).
		Do(context.Background())
	if err != nil {
		log.Println("createMarketOrder - 2:", err)
		return
	}
	log.Println("STOP_MARKET Order:", order)

	// 新建市价单
	order, err = client.NewCreateOrderService().Symbol(cfg.Symbol).
		Side(sideType).Type("MARKET").
		Quantity(utils.F64ToStr(qty)).
		Do(context.Background())
	if err != nil {
		log.Println("createMarketOrder - 3:", err)
		return
	}
	log.Println("MARKET Order:", order)

}

func (b *TradeSrv) ClosePosition(posAmt float64) {
	if posAmt == 0 {
		return
	}
	qty := posAmt

	sideType := futures.SideTypeSell
	if posAmt < 0 {
		sideType = futures.SideTypeBuy
		qty = math.Abs(posAmt)
	}

	order, err := client.NewCreateOrderService().Symbol(cfg.Symbol).
		Side(sideType).Type("MARKET").
		Quantity(utils.F64ToStr(qty)).
		Do(context.Background())
	if err != nil {
		log.Println("closePosition1", err)
		return
	}

	log.Println("ClosePosition:", order)

	err = client.NewCancelAllOpenOrdersService().Symbol(cfg.Symbol).Do(context.Background())
	if err != nil {
		log.Println("closePosition2", err)
		return
	}
}
