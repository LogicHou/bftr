package webhdl

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/LogicHou/bftr/store"
)

func testHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi, from Service: " + name))
	}
}

func InfoHandler(td *store.Trader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		if r.Form["pwd"] != nil && r.Form["pwd"][0] == "ss3CAK4K3KYQi" {
			var issueList = template.Must(template.New("issuelist").Parse(`
<table border="1" style="
    width: 100%;
    height: 100%;
    font-size: 83px;
">
<tr style='text-align: left'>
  <th>Name</th>
  <th>Value</th>
</tr>
<tr>
  <td>Leverage</td>
  <td>{{.Leverage}}</td>
</tr>
<tr>
  <td>PosSide</td>
  <td>{{.PosSide}}</td>
</tr>
<tr>
  <td>EntryPrice</td>
  <td>{{.EntryPrice}}</td>
</tr>
<tr>
  <td>PosAmt</td>
  <td>{{.PosAmt}}</td>
</tr>
<tr>
  <td>PosQty</td>
  <td>{{.PosQty}}</td>
</tr>
<tr>
  <td>StopLoss</td>
  <td>{{.StopLoss}}</td>
</tr>
</table>
`))
			type SomeData struct {
				Leverage   float64
				PosSide    string
				EntryPrice float64
				PosAmt     float64
				PosQty     int
				StopLoss   float64
			}
			tpldata := SomeData{
				Leverage:   td.Leverage,
				PosSide:    string(td.PosSide),
				EntryPrice: td.EntryPrice,
				PosAmt:     td.PosAmt,
				PosQty:     td.PosQty,
				StopLoss:   td.StopLoss,
			}
			issueList.Execute(w, tpldata)

		} else {
			fmt.Fprintf(w, ":)")
		}
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
