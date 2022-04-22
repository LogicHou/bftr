package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/LogicHou/bftr/store/factory"
)

type Cfg struct {
	Symbol    string `json:"symbol"`
	ApiKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Interval  string `json:"interval"`
}

func main() {
	cfgFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err.Error())
		panic("config open error")
	}
	defer cfgFile.Close()

	byteCfgValue, _ := ioutil.ReadAll(cfgFile)

	var cfg Cfg
	json.Unmarshal([]byte(byteCfgValue), &cfg)
	fmt.Println(cfg)

	s, err := factory.New("mem")
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
