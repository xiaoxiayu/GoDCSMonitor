// rpc_server
package main

import (
	"fmt"
	"strconv"
	"strings"
	//	"path/filepath"
	"html/template"
	//	"log"
	"net/http"
	"net/rpc"
	//	"os"
	"regexp"
	"time"
)

type HttpMonitor struct {
	ScriptMonitor   string
	RpcMap          map[string]*rpc.Client
	ReConnectRpcMap map[string]bool
	GlobalCfg       Global
}

type home struct {
	Title string
}

type MonitorItem struct {
	Script string
	Pid    string
}

type TemplateData struct {
	Info     string
	Ip       string
	KillFunc template.JSStr
	InfoMap  map[string]string
	Platform string
}

func GetPidTemplate() string {
	return `
	<html>
	<pre>
	<div>
	<form method="post" action="kill" enctype="multipart/form-data">
	<p><input type="text" onchange="txt.value=this.value" class="file" name="pid"/></p>
	<p><input type="text" onchange="txt.value=this.value" class="file" name="ip"/></p>
	<p><input type="text" onchange="txt.value=this.value" class="file" name="platform"/></p>
	<input id="ii" type="submit" value="OK" />
	</p>
	</form>
	</div>
	</pre>
	</html>
`
}

func GetIndexTemplate() string {
	return `
	<html>
	<head>
	<script>
		function createXMLHttpRequest() {  
	    var xmlHttp;  
	    if (window.XMLHttpRequest) {  
	        xmlHttp = new XMLHttpRequest();  
	        if (xmlHttp.overrideMimeType)  
	            xmlHttp.overrideMimeType('text/xml');  
	    } else if (window.ActiveXObject) {  
	        try {  
	            xmlHttp = new ActiveXObject("Msxml2.XMLHTTP");  
	        } catch (e) {  
	            try {  
	                xmlHttp = new ActiveXObject("Microsoft.XMLHTTP");  
	            } catch (e) {  
	            }  
	        }  
	    }  
	    return xmlHttp;  
	} 
	function getStatusBack(result,HttpContents){
	 if (result.status==Http.Status.OK){
	  HttpContents.innerHTML = result.responseText;
	 } else {
	  HttpContents.innerHTML = "An error occurred (" + result.status.toString() + ").";
	 }
	}

	function check_os() {
		windows = (navigator.userAgent.indexOf("Windows",0) != -1)?1:0;
		mac = (navigator.userAgent.indexOf("mac",0) != -1)?1:0;
		linux = (navigator.userAgent.indexOf("Linux",0) != -1)?1:0;
		unix = (navigator.userAgent.indexOf("X11",0) != -1)?1:0;
	 
		if (windows) os_type = "Windows";
		else if (mac) os_type = "Mac";
		else if (linux) os_type = "Lunix";
		else if (unix) os_type = "Unix";
	 
		return os_type;
	}

	function post(path, params, method) {
	    method = method || "post"; // Set method to post by default if not specified.
	
	    // The rest of this code assumes you are not using a library.
	    // It can be made less wordy if you use one.
	    var form = document.createElement("form");
	    form.setAttribute("method", method);
	    form.setAttribute("action", path);
	
	    for(var key in params) {
	        if(params.hasOwnProperty(key)) {
	            var hiddenField = document.createElement("input");
	            hiddenField.setAttribute("type", "hidden");
	            hiddenField.setAttribute("name", key);
	            hiddenField.setAttribute("value", params[key]);
	
	            form.appendChild(hiddenField);
	         }
	    }
	
	    document.body.appendChild(form);
	    form.submit();
	}

	function kill(ip_id, info) {
		var objs = document.getElementById(ip_id);
		//alert(ip_id);
		//alert(info);
	
		post('/kill', {pid: info, ip: ip_id});
		alert("Kill: " + info + "@" + ip_id)
	}  

	</script>
	</head>
	<body>
		{{range $ip, $info:=.InfoMap}}
		</br>
			<a id="{{$ip}}">{{$ip}}:<p/>{{$info}}</a>
			<a><button type="button" onClick="kill({{$ip}}, {{$info}})()";>Kill</button></a>
			<a><p>--------------------------------</p></a>
		{{end}}
	</body>
	</html>
	`
}

type HtmlTempaleFunc struct {
	index int
}

func (htmlFunc HtmlTempaleFunc) ParseFuncName(funcName string) string {
	return strings.Replace(funcName, "\"", "", -1)
}

var mux map[string]func(http.ResponseWriter, *http.Request)

func (t *HttpMonitor) Index(w http.ResponseWriter, r *http.Request) {
	for _, Ip := range t.GlobalCfg.RemoteIP {
		if Ip.TaskMonitorPort == "" {
			Ip.TaskMonitorPort = "9093"
		}
		fmt.Println("== Monitor Try To Link: " + Ip.Value + ":" + Ip.TaskMonitorPort + " ==\n")
		client, err := rpc.Dial("tcp", Ip.Value+":"+Ip.TaskMonitorPort)
		if err != nil {
			fmt.Println("Monitor connect error:" + err.Error())
			fmt.Println("== Monitor Link ERROR: " + Ip.Value + ":" + Ip.TaskMonitorPort + " ==")
			continue
		}
		t.RpcMap[Ip.Value+":"+Ip.TaskMonitorPort] = client

		fmt.Println("== Monitor Link Success: " + Ip.Value + ":" + Ip.TaskMonitorPort + " ==\n")
	}

	for reconnect_ip, _ := range t.ReConnectRpcMap {
		client, err := rpc.Dial("tcp", reconnect_ip)
		if err != nil {
			fmt.Println("Monitor connect error:" + err.Error())
			continue
		}
		delete(t.ReConnectRpcMap, reconnect_ip)
		fmt.Println(t.ReConnectRpcMap)
		t.RpcMap[reconnect_ip] = client
	}

	var allInfo []byte
	p := TemplateData{Info: "asdfasdf", Ip: "a", KillFunc: "asdf"}
	p.InfoMap = make(map[string]string)

	for ip, rpcMonitor := range t.RpcMap {
		item := MonitorItem{Script: "asdf"}
		var remoteInfo []byte
		err := rpcMonitor.Call("RPC.GetRunningInfo", item, &remoteInfo)
		if err != nil {
			fmt.Println(string(err.Error()))
			if strings.Contains(err.Error(), "connection is shut down") {
				t.ReConnectRpcMap[ip] = true
				continue
			}
		}

		tem_str_list := strings.Split(string(remoteInfo), "_Platform_")
		fin_info := tem_str_list[0]

		tem_str_list = strings.Split(tem_str_list[1], "_ExternMonitorP_")
		platform_info := tem_str_list[0]

		//p.Platform = platform_info

		p.InfoMap[ip+"@"+platform_info] = fin_info

		if len(tem_str_list) > 1 {
			extern_monitor_str := tem_str_list[1]
			extern_monitor_list := strings.Split(extern_monitor_str, "\n")
			if platform_info == "windows" {

				if len(extern_monitor_list) >= 3 {
					for i, ch_info := range extern_monitor_list[3:] {
						tem_str := strings.Replace(ch_info, " ", "", -1)
						tem_str = strings.Replace(tem_str, "\n", "", -1)
						if len(tem_str) > 0 {
							p.InfoMap[ip+"-ExternProcess"+strconv.Itoa(i)+"@"+platform_info] = ch_info
							fmt.Println(ch_info)
						}
					}
				}
			} else {
				for i, ch_info := range extern_monitor_list {
					if strings.Contains(ch_info, "grep ") {
						continue
					}
					tem_str := strings.Replace(ch_info, " ", "", -1)
					tem_str = strings.Replace(tem_str, "\n", "", -1)
					ch_info = strings.Replace(ch_info, "/", "", -1)
					if len(tem_str) > 0 {
						p.InfoMap[ip+"-ExternProcess"+strconv.Itoa(i)+"@"+platform_info] = ch_info
					}
				}
			}
		}

		allInfo = append(allInfo, []byte(ip)...)
		allInfo = append(allInfo, []byte(":\t\n\t")...)
		allInfo = append(allInfo, remoteInfo...)
	}
	if len(p.InfoMap) > 0 {
		temp := template.New("index")
		temp, err := temp.Parse(GetIndexTemplate())
		if err != nil {
			fmt.Println(err.Error())
		}

		temp.Execute(w, p)
	} else {
		fmt.Fprintf(w, "Reflash")
	}
}

func (t *HttpMonitor) GetPid(w http.ResponseWriter, r *http.Request) {
	temp := template.New("GetPid")
	temp, err := temp.Parse(GetPidTemplate())
	if err != nil {
		fmt.Println(err.Error())
	}

	temp.Execute(w, r)
	fmt.Println("HTTP GetPID Run")
}

func (t *HttpMonitor) Kill(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	tem_list := strings.Split(string(r.Form["ip"][0]), "@")
	platform := tem_list[1]
	ip := strings.Split(tem_list[0], "-")[0]
	pid := r.Form["pid"][0]

	fmt.Println("PLATFORM", platform)

	if platform == "windows" {
		if strings.Contains(r.Form["ip"][0], "-ExternProcess") {
			tem_list := strings.Split(pid, " ")
			fmt.Println(tem_list[0])
			bin_name := tem_list[0]
			tem_list = strings.Split(pid, bin_name)
			tem_str := strings.TrimSpace(tem_list[1])
			fmt.Println(tem_str)
			tem_list2 := strings.Split(tem_str, " ")
			pid = tem_list2[0]
		} else {
			pid = strings.SplitAfterN(pid, "\t", -1)[4]
		}

	} else {
		pid = strings.Split(pid, " ")[0]
	}
	pid = strings.Replace(pid, "\t", "", -1)
	pid = strings.Replace(pid, "\n", "", -1)
	fmt.Println("======================")
	fmt.Println("*****", pid)

	item := MonitorItem{Pid: pid}
	var remoteInfo []byte
	//	fmt.Println(ip)
	//	fmt.Println("====================")
	err := t.RpcMap[ip].Call("RPC.KillProcess", item, &remoteInfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (*HttpMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		//fmt.Println(mux)
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		//fmt.Println("2")
		http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))).ServeHTTP(w, r)
	} else {
		//fmt.Println("3")
		http.StripPrefix("/", http.FileServer(http.Dir("."))).ServeHTTP(w, r)
	}
}

func (t *HttpMonitor) Init(cfg Global) {
	t.RpcMap = make(map[string]*rpc.Client)
	t.ReConnectRpcMap = make(map[string]bool)
	t.GlobalCfg = cfg

}

func StartHttpMonitor(cfg Global) {

	http_serv := new(HttpMonitor)
	http_serv.Init(cfg)

	server := http.Server{
		Addr:        ":" + cfg.MonitorPort,
		Handler:     &HttpMonitor{},
		ReadTimeout: 5 * time.Second,
	}

	mux = make(map[string]func(http.ResponseWriter, *http.Request))

	mux["/"] = http_serv.Index
	mux["/getpid"] = http_serv.GetPid
	mux["/kill"] = http_serv.Kill

	err := server.ListenAndServe()
	fmt.Println(err.Error())
}
