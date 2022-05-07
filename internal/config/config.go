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
	Symbol     string  `json:"symbol"`
	ApiKey     string  `json:"apiKey"`
	SecretKey  string  `json:"secretKey"`
	Interval   string  `json:"interval"`
	Leverage   string  `json:"leverage"`
	Margin     float64 `json:"margin"`
	OpenSideMa int     `json:"openSideMa"`
	CloseMa    int     `json:"closeMa"`
	OpenK1     float64 `json:"openK1"`
	OpenK2     float64 `json:"openK2"`
	OpenK3     float64 `json:"openK3"`
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
