package store

import (
	"sync"

	tstore "github.com/LogicHou/bftr/store"
	factory "github.com/LogicHou/bftr/store/factory"
)

func init() {
	factory.Register("mem", &MemStore{})
}

type MemStore struct {
	sync.RWMutex
	trader *tstore.Trader
}

// 市价下单函数 RestAPI
func (ms *MemStore) CreateMarketOrder(trader *tstore.Trader) error {

	return nil
}

// 平仓现有的所有仓位函数
func (ms *MemStore) ClosePosition(trader *tstore.Trader) error {

	return nil
}

// 可能用来刷新一些数据
func (ms *MemStore) RefreshData(trader *tstore.Trader) error {

	return nil
}
