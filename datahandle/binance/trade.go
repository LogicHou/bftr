package binance

import (
	"context"
	"log"
	"math"

	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

type TradeSrv struct{}

func NewTradeSrv() *TradeSrv {
	return &TradeSrv{}
}

func (b *TradeSrv) CreateMarketOrder() {

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
