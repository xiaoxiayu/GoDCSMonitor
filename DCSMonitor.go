// rpc_server
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"runtime"
	//	"time"
)

type Global struct {
	RemoteIP      []RemoteIP
	Workspace     string
	TransportPort string
	RunnerPort    string
	MonitorPort   string
	RuntimePath   string
}

type RemoteIP struct {
	RunnerPort      string `xml:"runnerport,attr"`
	TransportPort   string `xml:"transportport,attr"`
	TaskMonitorPort string `xml:"taskmonitorport,attr"`
	Value           string `xml:",chardata"`
}

var cfg_global Global

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	//	StartTime := time.Now().Format("2006-01-02 15:04:05")
	//	fmt.Println("")
	//	fmt.Println(StartTime)
	//	return
	xml_file := "globalcfg.xml"
	//fmt.Print(xml_file + "\n")
	file, err := ioutil.ReadFile(xml_file)
	if err != nil {
		fmt.Printf("%s Read ERROR: %v\n", xml_file, err)
	}
	//cfg_global = Global{}
	err = xml.Unmarshal(file, &cfg_global)
	if err != nil {
		fmt.Printf("%s XML Parse ERROR: %v\n", xml_file, err)
	}
	//fmt.Println(cfg_global)

	StartHttpMonitor(cfg_global)

}
