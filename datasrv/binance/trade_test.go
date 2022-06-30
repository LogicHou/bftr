package binance

import (
	"testing"
)

func TestCalcMqrginQty(t *testing.T) {
	tradeSrv := NewTradeSrv()
	qty, _ := tradeSrv.CalcMqrginQty(1027.53, 2)
	if qty != 0.212 {
		t.Errorf("not right")
	}
}
