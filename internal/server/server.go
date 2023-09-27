package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/scutrobotlab/asuwave/internal/helper"
	"github.com/scutrobotlab/asuwave/internal/variable"
)

// Start server
func Start(fsys *fs.FS) {
	port := ":" + strconv.Itoa(helper.Port)
	glog.Infoln("Listen on " + port)

	fmt.Println("asuwave running at:")
	fmt.Println("- Local:   http://localhost" + port + "/")
	ips := getLocalIP()
	for _, ip := range ips {
		fmt.Println("- Network: http://" + ip + port + "/")
	}
	fmt.Println("Don't close this before you have done")

	variableToReadCtrl := makeVariableCtrl(variable.RD)
	variableToWriteCtrl := makeVariableCtrl(variable.WR)
	//为.js扩展名添加MIME类型，这样服务器可以正确地提供JavaScript文件。
	mime.AddExtensionType(".js", "application/javascript")
	//配置文件服务器以提供在fsys中的文件
	http.Handle("/", http.FileServer(http.FS(*fsys)))

	//设置不同的HTTP路由和对应的控制器
	http.Handle("/serial", logs(serialCtrl))
	http.Handle("/serial_cur", logs(serialCurCtrl))
	http.Handle("/variable_read", logs(variableToReadCtrl))
	http.Handle("/variable_write", logs(variableToWriteCtrl))
	http.Handle("/variable_proj", logs(variableToProjCtrl))
	http.Handle("/variable_type", logs(variableTypeCtrl))
	http.Handle("/file/upload", logs(fileUploadCtrl))
	http.Handle("/file/path", logs(filePathCtrl))
	http.Handle("/option", logs(optionCtrl))
	http.Handle("/dataws", logs(dataWebsocketCtrl))
	http.Handle("/filews", logs(fileWebsocketCtrl))
	//启动HTTP服务器并监听之前定义的端口.如果出现错误，则打印错误日志并结束程序。
	glog.Fatalln(http.ListenAndServe(port, nil))
}

// 对任何传入的HTTP处理函数增加日志记录功能。当处理一个HTTP请求时，它会先打印请求的远程地址、方法和URL，然后再调用原始的处理函数。
func logs(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infoln(r.RemoteAddr, r.Method, r.URL)
		http.HandlerFunc(f).ServeHTTP(w, r)
	})
}

func errorJson(s string) string {
	j := struct{ Error string }{s}
	b, _ := json.Marshal(j)
	return string(b)
}

func getLocalIP() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	return ips
}
