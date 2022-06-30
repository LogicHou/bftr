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
}

func (srv *Server) Serve() error {
	resL, err := client.NewChangeLeverageService().Symbol(cfg.Symbol).Leverage(int(cfg.Leverage)).Do(context.Background())
	if err != nil || float64(resL.Leverage) != cfg.Leverage {
		return errors.Errorf("change leverage failed: %s", err)
	}
	for {
		doneC, _, err := futures.WsKlineServe(cfg.Binance.Symbol, cfg.Binance.Interval,
			func(event *futures.WsKlineEvent) {
				srv.WskChan <- event
			},
			func(err error) {
				err = errors.Errorf(time.Now().Format("2006-01-02 15:04:05"), "errmsg: %s", err)
			})
		if err != nil {
			return errors.Errorf(time.Now().Format("2006-01-02 15:04:05"), "doneC err: %s", err)
		}
		<-doneC
	}
}
