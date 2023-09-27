package server

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"

	"github.com/scutrobotlab/asuwave/internal/serial"
	"github.com/scutrobotlab/asuwave/internal/variable"
)

// makeVariableCtrl 接受一个 variable.Mod 类型参数，并返回一个用于控制变量的HTTP处理函数。
// vList 要控制的参数列表；
// isVToRead 为true代表只读变量，为false代表可写变量。
func makeVariableCtrl(m variable.Mod) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// 保证请求的Body在函数结束时关闭，以防资源泄露。
		defer r.Body.Close()

		// 设置响应的内容类型为JSON。
		w.Header().Set("Content-Type", "application/json")

		var err error

		// 根据请求的HTTP方法进行处理。
		switch r.Method {
		// 当请求方法为GET时，获取并返回所有变量。
		case http.MethodGet:
			b, _ := variable.GetAll(m)   // 获取所有变量。
			io.WriteString(w, string(b)) // 将变量列表写入响应。

		// 当请求方法为POST时，添加新的变量。
		case http.MethodPost:
			var newVariable variable.T
			postData, _ := io.ReadAll(r.Body)            // 读取请求体。
			err = json.Unmarshal(postData, &newVariable) // 将请求体的JSON数据反序列化为变量结构。
			if err != nil {
				// 如果JSON数据无效，则返回400 Bad Request错误。
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, errorJson("Invaild json"))
				return
			}
			// 验证地址的有效性。
			if newVariable.Addr < 0x20000000 || newVariable.Addr >= 0x80000000 {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, errorJson("Address out of range"))
				return
			}
			// 检查地址是否已经被使用。
			if _, ok := variable.Get(m, newVariable.Addr); ok {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, errorJson("Address already used"))
				return
			}
			// 设置新变量。
			// 发送读命令。
			err = serial.SendCmd(variable.Subscribe, variable.ConvertTToCmdT(newVariable))
			variable.Set(m, newVariable.Addr, newVariable)
			w.WriteHeader(http.StatusNoContent) // 返回204 No Content响应。
			io.WriteString(w, "")

		// 当请求方法为PUT时，修改变量的值。
		case http.MethodPut:
			// 检查是否允许修改变量。
			if m == variable.RD {
				w.WriteHeader(http.StatusMethodNotAllowed) // 如果变量是只读的，则返回405 Method Not Allowed错误。
				io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
				return
			}
			var modVariable variable.T
			postData, _ := io.ReadAll(r.Body)            // 读取请求体。
			err = json.Unmarshal(postData, &modVariable) // 将请求体的JSON数据反序列化为变量结构。
			if err != nil {
				// 如果JSON数据无效，则返回400 Bad Request错误。
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, errorJson("Invaild json"))
				return
			}
			// 检查串口是否打开。
			if serial.SerialCur.Name == "" {
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, "Not allow when serial port closed.") // 如果串口是关闭的，则返回500 Internal Server Error。
				return
			}
			// 发送写命令。
			err = serial.SendWriteCmd(modVariable)
			if err != nil {
				// 如果写命令失败，则返回500 Internal Server Error。
				w.WriteHeader(http.StatusInternalServerError)
				io.WriteString(w, errorJson(err.Error()))
				return
			}
			w.WriteHeader(http.StatusNoContent) // 返回204 No Content响应。
			io.WriteString(w, "")
		// 删除变量
		case http.MethodDelete:
			var oldVariable variable.T                   // 创建一个 variable.T 类型的变量来存储待删除的变量信息。
			postData, _ := io.ReadAll(r.Body)            // 读取请求体的数据。
			err = json.Unmarshal(postData, &oldVariable) // 将请求体的JSON数据反序列化到 oldVariable 中。
			if err != nil {
				w.WriteHeader(http.StatusBadRequest) // 如果JSON数据无效，则返回400 Bad Request错误。
				io.WriteString(w, errorJson("Invaild json"))
				return
			}

			// 之前的代码可能是用于检查指定地址的变量是否存在。但你认为这个检查是不必要的。
			// if _, ok := vList.Variables[oldVariable.Addr]; !ok {
			// 	w.WriteHeader(http.StatusBadRequest)
			// 	io.WriteString(w, errorJson("No such address"))
			// }

			variable.Delete(m, oldVariable.Addr)
			w.WriteHeader(http.StatusNoContent) // 返回204 No Content响应，表示请求已成功处理，但没有内容返回。
			io.WriteString(w, "")
			return // 结束此case，返回。

		default:
			w.WriteHeader(http.StatusMethodNotAllowed) // 对于不支持的HTTP方法，返回405 Method Not Allowed错误。
			io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
		}
	}
}

// 工程变量
func variableToProjCtrl(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		b, _ := variable.GetAllProj()
		io.WriteString(w, string(b))
	case http.MethodDelete:
		variable.SetAllProj(variable.Projs{})

		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}

func variableTypeCtrl(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		var types struct{ Types []string }
		for k := range variable.TypeLen {
			types.Types = append(types.Types, k)
		}
		sort.Strings(types.Types)
		b, _ := json.Marshal(types)
		io.WriteString(w, string(b))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, errorJson(http.StatusText(http.StatusMethodNotAllowed)))
	}
}
