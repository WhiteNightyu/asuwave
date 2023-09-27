package option

import (
	"flag"
	"os"
	"path"
	"strconv"

	"github.com/golang/glog"
	"github.com/scutrobotlab/asuwave/internal/helper"
	"github.com/scutrobotlab/asuwave/internal/variable"
	"github.com/scutrobotlab/asuwave/pkg/elffile"
	"github.com/scutrobotlab/asuwave/pkg/jsonfile"
)

var (
	logLevel     int  //log级别
	saveFilePath bool //保存文件路径的开关状态
)

var (
	optionPath    = path.Join(helper.AppConfigDir(), "option.json")    //程序的配置文件夹下的option.json
	fileWatchPath = path.Join(helper.AppConfigDir(), "FileWatch.json") //程序的配置文件夹下的FileWatch.json
)

type OptT struct {
	LogLevel     int
	SaveFilePath bool
	SaveVarList  bool
	UpdateByProj bool
}

func Get() OptT {
	return OptT{
		LogLevel:     logLevel,
		SaveFilePath: saveFilePath,
		SaveVarList:  variable.GetOptSaveVarList(),
		UpdateByProj: variable.GetOptUpdateByProj(),
	}
}

func Load() {
	var opt OptT
	jsonfile.Load(optionPath, &opt) //从optionPath中加载配置选项到opt
	//将opt中的SaveVarList和UpdateByProj配置选项分别设置到相关变量中
	variable.SetOptSaveVarList(opt.SaveVarList)
	variable.SetOptUpdateByProj(opt.UpdateByProj)

	var watchList []string
	jsonfile.Load(fileWatchPath, &watchList) //加载指定路径fileWatchPath的文件监视列表到字符串切片watchList中
	for _, w := range watchList {
		elffile.ChFileWatch <- w //将watchList中的每一个文件名发送到elffile.ChFileWatch通道中
	}
	//保存当前的文件监视列表和配置选项到对应的文件中
	jsonfile.Save(fileWatchPath, elffile.GetWatchList())
	jsonfile.Save(optionPath, opt)
}

func SetLogLevel(v int) {
	if logLevel == v {
		glog.V(1).Infof("LogLevel has set to %d, skip\n", v)
		return
	}
	glog.V(1).Infof("Set LogLevel to %d\n", v)
	logLevel = v
	if err := flag.Set("v", strconv.Itoa(v)); err != nil {
		glog.Errorln(err.Error())
	}
	jsonfile.Save(optionPath, Get())
}

func SetSaveFilePath(v bool) {
	if saveFilePath == v {
		glog.V(1).Infof("SaveFilePath has set to %t, skip\n", v)
		return
	}
	glog.V(1).Infof("Set SaveFilePath to %t\n", v)
	if v {
		jsonfile.Save(fileWatchPath, elffile.GetWatchList())
	} else {
		os.Remove(fileWatchPath)
	}
	saveFilePath = v
	jsonfile.Save(optionPath, Get())
}

func SetSaveVarList(v bool) {
	variable.SetOptSaveVarList(v)
	jsonfile.Save(optionPath, Get())
}

func SetUpdateByProj(v bool) {
	variable.SetOptUpdateByProj(v)
	jsonfile.Save(optionPath, Get())
}
