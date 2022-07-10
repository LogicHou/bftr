package webhdl

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LogicHou/bftr/store"
)

func testHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi, from Service: " + name))
	}
}

func KlineUpdateHdl(td *store.Trader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type kline struct {
			Timestamp int64   `json:"timestamp"`
			Open      float64 `json:"open"`
			High      float64 `json:"high"`
			Low       float64 `json:"low"`
			Close     float64 `json:"close"`
			Volume    float64 `json:"volume"`
		}
		data := kline{
			Timestamp: td.Wsk.ST,
			Open:      td.Wsk.O,
			High:      td.Wsk.H,
			Low:       td.Wsk.L,
			Close:     td.Wsk.C,
			Volume:    td.Wsk.V,
		}
		res, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

}

func closeHandler(w http.ResponseWriter, r *http.Request) {

}

func reHandler(w http.ResponseWriter, r *http.Request) {

}
