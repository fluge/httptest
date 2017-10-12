package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var jsonDataMap map[string]interface{}
var isShow bool

func main() {
	var port = flag.Int("p", 8090, "local httptest server port")
	var path = flag.String("path", "./test-fixtures/", "local fixtures path")
	var style = flag.String("f", "json", "fixtures style")
	var showLog = flag.Bool("s", false, "local httptest show log")
	flag.Parse()
	filePath := *path + *style
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		log.Println(err, "please use httptest -path to set fixtures path")
		return
	}
	isShow = *showLog
	jsonDataMap = make(map[string]interface{}, len(files))
	for _, v := range files {
		ret, err := readFile(filePath + "/" + v.Name())
		if err != nil {
			panic(err)
		}
		jsonDataMap[v.Name()] = ret
	}
	//对文件进行监控，如果有修改就自动重新加载文件
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	//添加监控文件
	err = watcher.Add(filePath + "/")
	if err != nil {
		log.Println(err)
		return
	}
	go func() { //监控文件变化，如果文件改变可以重新加载改变的文件
		for {
			select {
			case ev := <-watcher.Events:
				if (ev.Op&fsnotify.Create == fsnotify.Create || ev.Op&fsnotify.Write == fsnotify.Write) && !strings.Contains(ev.Name, "jb_") {
					temp, err := readFile("./" + ev.Name)
					if err == nil {
						jsonDataMap[ev.Name] = temp
					}
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}()
	//读取配置文件，获取端口号
	ports := fmt.Sprintf(":%d", *port)
	log.Println("http test server start at", ports)
	http.HandleFunc("/", Handle)
	err = http.ListenAndServe(ports, nil)
	if err != nil {
		log.Println(err)
	}
}

//对请求进行处理
func Handle(w http.ResponseWriter, r *http.Request) {
	var res map[string]interface{}
	host, urlP := splitPath(r.URL.Path)
	if isShow {
		log.Println("path:", host, urlP, r.Method)
	}
	fileName := host + ".json"
	if temp, ok := jsonDataMap[fileName]; ok {
		res = temp.(map[string]interface{})
	} else {
		log.Println("no " + host + " exist")
		return
	}
	//读取request 中的参数,判断请求的方式
	arr := ""
	switch r.Method {
	case "GET":
		//对get请求中的参数进行匹配
		pathArr := strings.Split(r.RequestURI, "?")
		if len(pathArr) > 1 {
			arr = pathArr[1]
		}
	case "POST":
		//对post请求进行处理
		date, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		arr = string(date)
		if strings.Contains(arr, "{") { //认为是json请求，就拼凑请求中的内容
			arr = json2Str(arr)
		}
	default:
		w.Write([]byte("unsupported " + r.Method + " method"))
	}
	sli := strings.Split(arr, "&")
	key := make([]string, 0)
	for _, v := range sli { //对请求体中的参数进行过滤
		if !(strings.Contains(v, "signature") || strings.Contains(v, "time") || strings.Contains(v, "appkey") || strings.Contains(v, "expires") || strings.Contains(v, "nonce")) {
			key = append(key, v)
		}
	}
	keys := strings.Join(key, "&")
	if isShow {
		log.Println("key:", fmt.Sprintf("%v", keys))

	}
	//对fixtures中的参数进行匹配，并返回对应的response
	if temp, ok := res[urlP]; ok { //判断时候是否是*
		jsonToResponse(w, temp.(map[string]interface{}), keys)
	} else {
		w.Write([]byte("no match url ,url should is : " + urlP))
	}
}

//解析json文件
func jsonToResponse(w http.ResponseWriter, arr map[string]interface{}, keys string) {
	//遍历匹配key
	for k, v := range arr {
		isKey := true
		if k != keys { //如果不能直接匹配就判断包含关系
			arr := strings.Split(k, "&")
			for _, v := range arr {
				if !strings.Contains(keys, v) {
					isKey = false
					break
				}
			}
		}
		if isKey {
			date := map[string]interface{}{
				"code":    0,
				"message": "succeed",
				"data":    v,
			}
			data, err := json.Marshal(date)
			if err != nil {
				panic(err)
			}
			if isShow {
				log.Println("response :", string(data))
			}
			w.Write(data)
			return
		}
	}
	if isShow {
		log.Println("no match key,key should is :" + keys)

	}
}

func readFile(fileName string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("read file err: ", err.Error())
		return nil, err
	}

	var ret map[string]interface{}
	if err := json.Unmarshal(bytes, &ret); err != nil {
		log.Println("read file json unmarshal: ", err.Error())
		return nil, err
	}

	return ret, nil
}

func splitPath(path string) (string, string) {
	arr := strings.Split(path, "/")[1:]
	s := ""
	p := ""
	for k, v := range arr {
		if k == 0 {
			s = v
		} else {
			p = p + "/" + v
		}
	}
	return s, p
}

func json2Str(jsonData string) string {
	arr := make([]rune, 0)
	isT := false
	for _, v := range jsonData {
		if v == int32(34) || v == int32(123) || v == int32(125) {
			continue
		}
		if v == int32(58) {
			v = int32(61)
		}
		if v == int32(44) {
			v = int32(38)
		}
		if v == int32(92) {
			isT = true
		}
		arr = append(arr, v)
	}
	str := string(arr)
	//处理转义
	if isT {
		str = strings.Replace(str, `\u003c`, "<", -1)
		str = strings.Replace(str, `\u003e`, ">", -1)
		str = strings.Replace(str, `\u0026`, "&", -1)
	}
	return str
}
