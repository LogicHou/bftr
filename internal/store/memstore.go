package store

import (
	"sync"

	"github.com/LogicHou/bftr/indicator"
	tstore "github.com/LogicHou/bftr/store"
	factory "github.com/LogicHou/bftr/store/factory"
	"github.com/adshao/go-binance/v2/futures"
)

func init() {
	factory.Register("mem", &MemStore{
		trader: &tstore.Trader{
			HistKlines:  nil,
			PosAmt:      0, // 持仓金额，值等于0的时候表示未持仓
			PosQty:      0, // 持仓K次数
			EntryPrice:  0, // 开仓均价
			Openk:       0,
			PosSide:     futures.SideTypeBuy, // 持仓的买卖方向，默认为买
			StopLoss:    0,                   // 止损数值
			RefreshTime: map[string]int64{"4h": 14404000, "1h": 3604000, "30m": 1804000, "15m": 904000, "5m": 304000, "1m": 64000},
		}})
}

type MemStore struct {
	sync.RWMutex
	trader *tstore.Trader
}

func (ms *MemStore) Get() *tstore.Trader {
	ms.RLock()
	defer ms.RUnlock()

	return ms.trader
}

func (ms *MemStore) Update(klines []*indicator.Kline) {
	ms.Lock()
	defer ms.Unlock()

	ms.trader.HistKlines = klines
}

func (ms *MemStore) Reset() error {
	ms.Lock()
	defer ms.Unlock()

	ms.trader.StopLoss = 0
	ms.trader.PosQty = 0
	ms.trader.Openk = 0
	ms.trader.PosSide = futures.SideTypeBuy

	return nil
}
