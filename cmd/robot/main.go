package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/LogicHou/bftr/internal/store"
	"github.com/LogicHou/bftr/server"
	"github.com/LogicHou/bftr/store/factory"
)

func main() {
	var logToFile int
	flag.IntVar(&logToFile, "lf", 0, "log to file")
	flag.Parse()

	if logToFile == 1 {
		logFile, err := os.OpenFile(`./bftr.log`, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		log.SetOutput(logFile)
	}

	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	srv := server.NewTradeServer(s)
	sErrChan, err := srv.ListenAndServe()
	if err != nil {
		log.Println("trader server start failed:", err)
		return
	}
	log.Println("trader server start ok")

	mErrChan, err := srv.ListenAndMonitor()
	if err != nil {
		log.Println("trader monitor start failed:", err)
		return
	}
	log.Println("trader monitor start ok")

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-sErrChan:
		log.Println("trader server have some errors:", err)
		return
	case err = <-mErrChan:
		log.Println("trader monitor have some errors:", err)
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
