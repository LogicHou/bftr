package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func init() {

}

type Cfg struct {
	Binance `json:"binance"`
}

type Binance struct {
	Symbol    string `json:"symbol"`
	ApiKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Interval  string `json:"interval"`
}

func Get() *Cfg {
	workDir, _ := os.Getwd()
	cfgFile, err := os.Open(workDir + "/config.json")
	if err != nil {
		fmt.Println(err.Error())
		panic("config open error")
	}
	defer cfgFile.Close()

	byteCfgValue, _ := ioutil.ReadAll(cfgFile)

	var cfg Cfg
	json.Unmarshal([]byte(byteCfgValue), &cfg)
	return &cfg
}
