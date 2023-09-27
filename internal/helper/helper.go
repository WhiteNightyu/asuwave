package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// 定义全局变量，存储程序的版本信息和编译时间等信息
var (
	Port      int
	GitTag    string
	GitHash   string
	BuildTime string //编译时间
	GoVersion string //GO版本
)

// 定义Github发布信息的结构
type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// init函数在程序启动时执行，用于检查程序的配置文件夹是否存在，如果不存在则创建它
func init() {
	//os.Stat(): 返回指定路径的文件或目录的FileInfo（包括文件名、大小、权限等属性）和一个错误值err。
	//os.IsNotExist(err): 用于检查错误值err是否表示“文件或目录不存在”
	if _, err := os.Stat(AppConfigDir()); os.IsNotExist(err) {
		//os.MkdirAll()用于创建指定路径的目录，包括所有必要的父级目录。如果目录已经存在，则不会进行任何操作，也不会返回错误。
		//0755，八进制数，表示目录的权限为 rwxr-xr-x，即用户读、写、执行权限，组和其他用户只有读和执行权限。
		err := os.MkdirAll(AppConfigDir(), 0755)
		if err != nil {
			panic(err) //在遇到无法恢复的错误或异常情况时中断正常的程序执行
		}
	}
}

// AppConfigDir返回程序的配置文件夹路径
func AppConfigDir() string {
	dir, err := os.UserConfigDir() // 获取用户的配置文件夹路径
	if err != nil {                //如果用户文件夹获取错误，则将配置文件的路径选择在exe文件所在路径
		dir = "./"
	}

	return path.Join(dir, "asuwave") // 将路径与"asuwave"组合并返回
}

// GetVersion返回程序的版本信息
func GetVersion() string {
	return fmt.Sprintf("asuwave %s\nbuild time %s\n%s", GitHash, BuildTime, GoVersion)
}

// CheckUpdate用于检查Github上是否有新的版本发布
func CheckUpdate(auto bool) {
	// 发送GET请求到Github的API，获取最新的发布信息
	resp, err := http.Get("https://api.github.com/repos/scutrobotlab/asuwave/releases/latest")
	if err != nil {
		fmt.Println("network error: " + err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // 读取响应的内容
	var gr githubRelease
	// 将读取的内容解析为githubRelease结构
	if err := json.Unmarshal([]byte(body), &gr); err != nil {
		return
	}
	// 检查当前程序的版本是否是最新版本
	if GitTag == gr.TagName {
		fmt.Println("already the latest version: " + GitTag)
		return
	}
	// 遍历Github发布的资源，查找与当前操作系统和架构匹配的资源
	for _, asset := range gr.Assets {
		if strings.Contains(asset.Name, runtime.GOOS+"_"+runtime.GOARCH) {
			fmt.Println("current version is " + GitTag)
			fmt.Println("new version available: " + gr.TagName)
			fmt.Print("download now? (y/n) ")
			var a string
			// 如果auto为false，则等待用户输入是否下载新版本
			if !auto {
				fmt.Scanln(&a)
			}
			// 如果用户选择下载或auto为true，则开始下载新版本
			if auto || a == "y" || a == "Y" || a == "yes" {
				if err := DownloadFile(asset.BrowserDownloadURL, asset.Name); err != nil {
					// 如果下载失败，尝试从fastgit镜像站点下载
					fmt.Println("download error: " + err.Error())
					fmt.Println("trying hub.fastgit.org...")
					asset.BrowserDownloadURL = strings.Replace(asset.BrowserDownloadURL, "https://github.com", "https://hub.fastgit.org", 1)
					DownloadFile(asset.BrowserDownloadURL, asset.Name)
				}
			}
			return
		}
	}
	fmt.Printf("don't know your platform: %s, %s", runtime.GOOS, runtime.GOARCH)
}

func StartBrowser(url string) {
	//commands 是一个映射（map），将操作系统的名称作为键，对应的打开浏览器命令作为值
	var commands = map[string]string{
		"windows": "explorer.exe", //对于 Windows 系统，使用 "explorer.exe" 命令打开浏览器
		"darwin":  "open",         //对于 macOS 系统，使用 "open" 命令打开浏览器
		"linux":   "xdg-open",     //对于 Linux 系统，使用 "xdg-open" 命令打开浏览器
	}
	//runtime.GOOS是当前操作系统的标识符
	run, ok := commands[runtime.GOOS]
	if !ok { //如果commands中没有标识符runtime.GOOS，即系统不是Windows、macOS、Linux
		fmt.Printf("don't know how to open things on %s platform", runtime.GOOS)
	} else {
		go func() {
			fmt.Println("Your browser will start")
			//创建一个外部命令，并启动该命令的执行
			//run表示要执行的命令或可执行文件的名称
			//url表示要传递给命令的参数
			exec.Command(run, url).Start()
		}()
	}
}
