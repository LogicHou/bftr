package datahandle

type Client interface {
	Start() error
}

type Kline interface {
	Get()
}

type Trade interface {
	Create()
	StopLoss()
}
