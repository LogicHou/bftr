package datahandle

import "github.com/LogicHou/bftr/indicator"

type Kline interface {
	Get(limit int) ([]*indicator.Kline, error)
}

type Trade interface {
	Create()
	StopLoss()
}
