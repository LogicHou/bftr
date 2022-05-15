package binance

import (
	"context"
	"log"
	"math"

	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
)

type TradeSrv struct {
	Margin       float64
	Interval     string
	Leverage     float64
	OpenSideMa   int
	PosQtyUlimit int
	CloseMa      int
	OpenK1       float64
	OpenK2       float64
	OpenK3       float64
}

func NewTradeSrv() *TradeSrv {
	return &TradeSrv{
		Margin:       cfg.Margin,
		Interval:     cfg.Interval,
		Leverage:     cfg.Leverage,
		OpenSideMa:   cfg.OpenSideMa,
		PosQtyUlimit: cfg.PosQtyUlimit,
		CloseMa:      cfg.CloseMa,
		OpenK1:       cfg.OpenK1,
		OpenK2:       cfg.OpenK2,
		OpenK3:       cfg.OpenK3,
	}
}

func (t *TradeSrv) CalcMqrginQty(curClose float64) float64 {
	if t.Margin > 0 {
		return utils.FRound(t.Margin * t.Leverage / curClose)
	}

	return 0
}

func (t *TradeSrv) CreateMarketOrder(sideType futures.SideType, qty float64, maxStopLoss float64) error {
	// 取消所有挂单
	err := client.NewCancelAllOpenOrdersService().Symbol(cfg.Symbol).Do(context.Background())
	if err != nil {
		return errors.Errorf("createMarketOrder - 1: %s", err)
	}
	sideStop := futures.SideTypeBuy
	offset := +5.0
	if sideType == futures.SideTypeBuy {
		sideStop = futures.SideTypeSell
		offset = -5.0
	}

	// 预埋止损单 RestAPI
	order, err := client.NewCreateOrderService().Symbol(cfg.Symbol).
		Side(sideStop).Type("STOP_MARKET").
		ClosePosition(true).StopPrice(utils.F64ToStr(utils.FRound2(maxStopLoss + offset))).
		Do(context.Background())
	if err != nil {
		return errors.Errorf("createMarketOrder - 2: %s", err)
	}
	log.Println("STOP_MARKET Order:", order)

	// 新建市价单
	order, err = client.NewCreateOrderService().Symbol(cfg.Symbol).
		Side(sideType).Type("MARKET").
		Quantity(utils.F64ToStr(qty)).
		Do(context.Background())
	if err != nil {
		err2 := client.NewCancelAllOpenOrdersService().Symbol(cfg.Symbol).Do(context.Background())
		return errors.Errorf("createMarketOrder - 3: %s %s", err, err2)
	}
	log.Println("MARKET Order:", order)

	return nil
}

func (t *TradeSrv) ClosePosition(posAmt float64) error {
	if posAmt == 0 {
		return errors.Errorf("posAmt is zero")
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
		return errors.Errorf("closePosition - 1: %s", err)
	}

	log.Println("ClosePosition:", order)

	err = client.NewCancelAllOpenOrdersService().Symbol(cfg.Symbol).Do(context.Background())
	if err != nil {
		return errors.Errorf("closePosition2: %s", err)
	}

	return nil
}

// 获取持仓信息 RestAPI
// 简言之, 您应该通过相关的rest接口( GET /fapi/v2/account 和 GET /fapi/v2/positionRisk) 获取资产和头寸的全量信息;
// 通过Websocket USER-DATA-STREAM 中的事件ACCOUNT_UPDATE对本地缓存的资产或头寸数据进行增量更新。
// https://binance-docs.github.io/apidocs/futures/cn/#v2-user_data-2
// https://binance-docs.github.io/apidocs/futures/cn/#v2-user_data-3
func (t *TradeSrv) GetPostionRisk() (posAmt float64, entryPrice float64, leverage float64, posSide futures.SideType, err error) {
	res, err := client.NewGetPositionRiskService().Symbol(cfg.Symbol).Do(context.Background())
	if err != nil {
		return
	}
	posAmt = utils.StrToF64(res[0].PositionAmt)
	entryPrice = utils.StrToF64(res[0].EntryPrice)
	leverage = utils.StrToF64(res[0].Leverage)

	if posAmt > 0 {
		posSide = futures.SideTypeBuy
	}
	if posAmt < 0 {
		posSide = futures.SideTypeSell
	}
	return
}
