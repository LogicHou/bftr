package indicator

type Kline struct {
	OpenTime  int64
	CloseTime int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	TradeNum  int64
	K         float64
	D         float64
	MA5       float64
	MA10      float64
	MA20      float64
	MA100     float64
}
