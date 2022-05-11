package binance

import (
	"context"
	"time"

	"github.com/LogicHou/bftr/internal/config"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"
)

var (
	cfg    *config.Cfg
	client *futures.Client
)

func init() {
	cfg = config.Get()
	client = binance.NewFuturesClient(cfg.Binance.ApiKey, cfg.Binance.SecretKey)
	client.NewSetServerTimeService().Do(context.Background())
}

type Server struct {
	WskChan chan *futures.WsKlineEvent
	// DoneChan chan struct{}
	// StopChan chan struct{}
}

func (srv *Server) Serve() error {
	for {
		doneC, _, err := futures.WsKlineServe(cfg.Binance.Symbol, cfg.Binance.Interval,
			func(event *futures.WsKlineEvent) {
				srv.WskChan <- event
			},
			func(err error) {
				err = errors.Errorf(time.Now().Format("2006-01-02 15:04:05"), "errmsg:", err)
			})
		if err != nil {
			return errors.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: ", err)
		}
		<-doneC
	}
}
