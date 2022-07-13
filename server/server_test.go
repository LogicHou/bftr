// calc_test.go
package server

import (
	"fmt"
	"testing"

	bds "github.com/LogicHou/bftr/datasrv/binance"
	"github.com/LogicHou/bftr/utils"
	"github.com/adshao/go-binance/v2/futures"
)

func TestOpenCondition(t *testing.T) {
	klineSrv := bds.NewKlineSrv()
	histKlines, _ := klineSrv.Get()
	klineSrv.WithKdj(histKlines)
	fmt.Println(utils.MsToTime(histKlines[len(histKlines)-1].OpenTime))
	fmt.Println(utils.MsToTime(histKlines[len(histKlines)-7].OpenTime))
	histKlines = histKlines[:len(histKlines)-7]

	srv := &TradeServer{
		srv: &bds.Server{
			WskChan: make(chan *futures.WsKlineEvent),
		},
		tradeSrv: bds.NewTradeSrv(),
	}
	r := srv.openCondition(futures.SideTypeSell, 38, histKlines)
	fmt.Println(r)
}
