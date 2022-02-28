package store

import (
	"errors"

	"github.com/adshao/go-binance/v2/futures"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExist    = errors.New("exist")
)

type Trader struct {
	pid
	wsk
}

// 一些负责控制下单平仓的控制参数
type pid struct {
	margin      float64
	marginRatio float64
	maxVARS5    float64
	maxVARS4    float64

	posAmt         float64
	entryPrice     float64
	leverage       float64
	posSide        futures.SideType
	stopLoss       float64
	firstStopLoss  float64
	gainQty        int
	interval       string
	openVolume     float64
	initOV         float64
	closeMA        int
	initCMA        int
	boxMarginLimit int
}

// websocket中ticket的详细数据
type wsk struct {
	h  float64
	l  float64
	c  float64
	v  float64
	E  int64
	cm float64
}

type Store interface {
}
