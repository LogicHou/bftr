package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var processName = "robot"
var processName_path = "./"

func findProcessID(processName string) (int, error) {
	buf := bytes.Buffer{}
	cmd := exec.Command("wmic", "process", "get", "name,processid")
	cmd.Stdout = &buf
	cmd.Run()

	cmd2 := exec.Command("findstr", processName)
	cmd2.Stdin = &buf
	data, _ := cmd2.CombinedOutput()
	if len(data) == 0 {
		return -1, errors.New("not find")
	}
	info := string(data)
	reg := regexp.MustCompile(`[0-9]+`)
	pid := reg.FindString(info)
	return strconv.Atoi(pid)
}

func getNowTime() string {
	timenow := time.Now().Format("2006-01-02 15:04:05")
	return timenow
}

func main() {
	pid, err := findProcessID(processName)
	if pid > 0 {
		fmt.Println("time:", getNowTime())
		fmt.Println(processName+" processID：", pid, " is runing")
	}
	if err == nil {
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("get process err:", err)
			return
		}
		process.Wait()
	}
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   os.Environ(),
	}
	for {
		_, err := findProcessID(processName)
		if err != nil {
			p, err := os.StartProcess(processName_path+processName, []string{processName_path+processName, "-lg", "log.txt"}, attr)
			if err != nil {
				fmt.Println("err:", err)
				return
			}
			fmt.Println("time:", getNowTime())
			fmt.Println(processName+" restart pid:", p.Pid)
			p.Wait()
			time.Sleep(1 * time.Second)
			fmt.Println("start ", processName)
		}
	}
}
