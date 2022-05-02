package strategy

import (
	"fmt"

	datasourcebn "github.com/LogicHou/bftr/datasource/binance"
)

type Order struct {
}

func (this *Order) Place() {
	KlineSrv := datasourcebn.KlineSrv{}
	err := KlineSrv.Get(145)
	if err != nil {
		fmt.Println(err)
		return
	}
	KlineSrv.WithKdj()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range KlineSrv.Klines {
		fmt.Printf("%.2f %.2f %.2f\n", v.Close, v.K, v.D)
		// fmt.Printf("%.2f %.2f %.2f\n", v.Close, utils.FRound2(k[i]), utils.FRound2(d[i]))
	}
}
