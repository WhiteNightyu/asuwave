package jsonfile

import (
	"encoding/json"
	"os"

	"github.com/golang/glog"
)

func Save(filename string, v interface{}) {
	jsonTxt, err := json.Marshal(v) //将v转化成JSON格式文本
	if err != nil {
		glog.Errorln(err.Error())
	}
	err = os.WriteFile(filename, jsonTxt, 0644)
	if err != nil {
		glog.Errorln(err.Error())
	}
	glog.Infoln(filename, "save success.")
}

func Load(filename string, v interface{}) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		glog.Infoln(filename, " unfound.")
		return
	}
	glog.Infoln(filename, "Found")
	data, err := os.ReadFile(filename)
	if err != nil {
		glog.Errorln(err.Error())
	}
	err = json.Unmarshal(data, v) //将读取到的数据解析到对象v中
	if err != nil {
		glog.Errorln(err.Error())
	}
}
