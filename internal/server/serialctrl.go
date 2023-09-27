package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/scutrobotlab/asuwave/internal/serial"
)

type SerialSetting struct {
	Serial string
	Baud   int
}

// serialCtrl 处理与序列相关的HTTP请求。
func serialCtrl(w http.ResponseWriter, r *http.Request) {
	// 保证请求的Body在函数结束时关闭，以防资源泄露。
	defer r.Body.Close()
	// 设置响应的内容类型为JSON。
	w.Header().Set("Content-Type", "application/json")

	// 根据请求的HTTP方法进行处理。
	switch r.Method {
	case http.MethodGet:
		// 当请求方法为GET时，获取所有的序列号并返回。
		j := struct{ Serials []string }{Serials: serial.Find()}
		b, _ := json.Marshal(j)      // 将结果转换为JSON格式。
		io.WriteString(w, string(b)) // 将JSON写入响应。

	default:
		// 如果请求方法不是GET，则返回405 Method Not Allowed错误。
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

// serialCurCtrl 处理与当前串口设置相关的HTTP请求。
func serialCurCtrl(w http.ResponseWriter, r *http.Request) {
	// 保证请求的Body在函数结束时关闭，以防资源泄露。
	defer r.Body.Close()

	// 设置响应的内容类型为JSON。
	w.Header().Set("Content-Type", "application/json")

	var err error

	// 根据请求的HTTP方法进行处理。
	switch r.Method {
	case http.MethodGet:
		// 当请求方法为GET时，获取当前的串口设置并返回。
		j := SerialSetting{
			Serial: serial.SerialCur.Name,          // 获取当前串口的名称。
			Baud:   serial.SerialCur.Mode.BaudRate, // 获取当前波特率。
		}
		b, _ := json.Marshal(j)      // 将结果转换为JSON格式。
		io.WriteString(w, string(b)) // 将JSON写入响应。

	case http.MethodPost:
		// 当请求方法为POST时，读取请求体中的JSON数据并尝试打开新的串口连接。
		j := SerialSetting{}
		postData, _ := io.ReadAll(r.Body)  // 读取请求体。
		err = json.Unmarshal(postData, &j) // 将请求体的JSON数据反序列化。
		if err != nil {
			// 如果JSON数据无效，则返回400 Bad Request错误。
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, errorJson("Invaild json"))
			return
		}

		// 尝试使用请求中提供的串口设置打开新的串口连接。
		err = serial.Open(j.Serial, j.Baud)
		if err != nil {
			// 如果打开串口失败，则返回500 Internal Server Error。
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		io.WriteString(w, string(postData)) // 将原始POST数据写入响应。

	case http.MethodDelete:
		// 当请求方法为DELETE时，尝试关闭当前的串口连接。
		err = serial.Close()
		if err != nil {
			// 如果关闭串口失败，则返回500 Internal Server Error。
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, errorJson(err.Error()))
			return
		}
		// 成功关闭串口后，返回204 No Content。
		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")

	default:
		// 如果请求方法不是GET、POST或DELETE，则返回405 Method Not Allowed错误。
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}
