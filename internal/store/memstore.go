package store

import (
	"sync"

	bds "github.com/LogicHou/bftr/datasrv/binance"
	tstore "github.com/LogicHou/bftr/store"
	factory "github.com/LogicHou/bftr/store/factory"
	"github.com/adshao/go-binance/v2/futures"
)

func init() {
	factory.Register("mem", &MemStore{
		trader: &tstore.Trader{
			Leverage:    1, // 当前杠杆倍数
			HistKlines:  nil,
			PosAmt:      0,                   // 持仓金额，值等于0的时候表示未持仓
			PosQty:      0,                   // 持仓K次数
			EntryPrice:  0,                   // 开仓均价
			PosSide:     futures.SideTypeBuy, // 持仓的买卖方向，默认为买
			StopLoss:    0,                   // 止损数值
			RefreshTime: map[string]int64{"30m": 3603000, "15m": 1803000, "5m": 603000},
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

func (ms *MemStore) Update() error {
	ms.Lock()
	defer ms.Unlock()

	var err error
	klineSrv := bds.NewKlineSrv()
	ms.trader.HistKlines, err = klineSrv.Get(41)
	if err != nil {
		return err
	}

	klineSrv.WithKdj(ms.trader.HistKlines)
	klineSrv.WithMa(ms.trader.HistKlines)

	return nil
}

func (ms *MemStore) Reset() error {
	ms.Lock()
	defer ms.Unlock()

	ms.trader.StopLoss = 0
	ms.trader.PosQty = 0
	ms.trader.PosSide = futures.SideTypeBuy
	return nil
}
