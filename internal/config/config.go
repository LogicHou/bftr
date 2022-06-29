package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func init() {

}

type Cfg struct {
	Binance `json:"binance"`
}

type Binance struct {
	Symbol       string  `json:"symbol"`
	ApiKey       string  `json:"apiKey"`
	SecretKey    string  `json:"secretKey"`
	Interval     string  `json:"interval"`
	Leverage     float64 `json:"leverage"`
	Margin       float64 `json:"margin"`
	MarginRatio  float64 `json:"marginRatio"`
	MarginLimit  float64 `json:"marginLimit"`
	HistRange    int     `json:"histRange"`
	OpenSideMa   int     `json:"openSideMa"`
	CloseMa      int     `json:"closeMa"`
	PosQtyUlimit int     `json:"posQtyUlimit"`
	OpenK1       float64 `json:"openK1"`
	OpenK2       float64 `json:"openK2"`
	OpenK3       float64 `json:"openK3"`
	KOffset      float64 `json:"kOffset"`
	CrossOffset  float64 `json:"crossOffset"`
}

func Get() *Cfg {
	workDir, _ := os.Getwd()
	cfgFile, err := os.Open(workDir + "/config.json")
	if err != nil {
		log.Println(err.Error())
		panic("config open error")
	}
	defer cfgFile.Close()

	byteCfgValue, _ := ioutil.ReadAll(cfgFile)

	var cfg Cfg
	json.Unmarshal([]byte(byteCfgValue), &cfg)
	return &cfg
}
