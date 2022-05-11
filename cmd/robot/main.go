package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/LogicHou/bftr/internal/store"
	"github.com/LogicHou/bftr/server"
	"github.com/LogicHou/bftr/store/factory"
)

func main() {
	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	srv := server.NewTradeServer(s)
	errChan := srv.ErrChan
	err = srv.Serve()
	if err != nil {
		log.Println("trader server start failed:", err)
		return
	}
	log.Println("trader server start ok")

	srv.Handler()
	log.Println("trader handler start ok")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select { // 监视来自errChan以及c的事件
	case err = <-errChan:
		log.Println("trader server run failed:", err)
		return
	case <-c:
		log.Println("program is exiting...")
		// @todo shutdown server
	}

	if err != nil {
		log.Println("program exit error:", err)
		return
	}
	log.Println("program exit ok")
}
