/*
1.获取index和diff数据
2.对请求文件进行sha1加密获取前10位做key到diff里查询
3.如果diff里查询到gitsha1直接访问uri/file/gitsha1值/文件名
4.如果diff里查询不到，计算目录sha1值(如xxx/yyyy/a.swf 计算 xxx/yyy 的sha1值)到index里查询，查询到对应gitsha1,访问cdnurl/tree/gitsha1/文件名
*/
package loadresource

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Loader struct {
	IsDebug bool
	uri     string
	index   map[string]string
	diff    map[string]string
}

func NewLoader(uri, baseTag, endTag string) (l *Loader, e1 error) {
	defer func() {
		if e := recover(); e != nil {

			switch e.(type) {
			case error:
				e1 = e.(error)
			default:
				e1 = fmt.Errorf("unknow error")
			}
		}
	}()

	l = &Loader{
		uri:   uri,
		index: getGitData(fmt.Sprintf("%sindex/%s", uri, baseTag)),
		diff:  getGitData(fmt.Sprintf("%sdiff/%s..%s", uri, baseTag, endTag)),
	}

	return
}

func (l *Loader) LoadData(file string) (contents []byte, e error) {
	f := fmt.Sprintf("%x", sha1.Sum([]byte(file)))[0:10]
	dirName, fileName := getFileInfo(file)

	if dir, ok := l.diff[f]; ok {
		uri := l.uri + "file/" + dir + "/" + fileName
		if l.IsDebug {
			println("load: ", uri)
		}
		contents, e = getFile(uri)
	} else {
		f = fmt.Sprintf("%x", sha1.Sum([]byte(dirName)))[0:10]

		var uri string
		if dir, ok = l.index[f]; ok {
			uri = l.uri + "tree/" + dir + "/" + fileName
		} else {
			uri = l.uri + "tree/" + fileName
		}

		if l.IsDebug {
			println("load: ", uri)
		}

		contents, e = getFile(uri)
	}

	return
}

func getGitData(uri string) map[string]string {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("read failed: %s", uri))
		}
	}()

	contents, e := getFile(uri)
	if e != nil {
		panic(e)
	}

	return formatGitData(fmt.Sprintf("%x", contents))
}

func getFile(uri string) ([]byte, error) {
	println(uri)
	println(strings.Index(uri, "http"))
	if strings.Index(uri, "http") > -1 {
		u, _ := url.Parse(uri)
		println("uri:", uri)
		res, _ := http.Get(u.String())
		return ioutil.ReadAll(res.Body)
	}

	return ioutil.ReadFile(uri)
}

func formatGitData(s string) map[string]string {
	n := 50

	result := make(map[string]string)
	if len(s) < n {
		return result
	}

	i := 0
	l := len(s)
	start := 0
	for {
		start = i * n
		result[s[start:start+10]] = s[start+10 : start+n]
		i++
		if i*n >= l {
			break
		}
	}
	return result
}

func getFileInfo(file string) (dir, fileName string) {
	arr := strings.Split(file, "/")
	l := len(arr)
	return strings.Join(arr[0:l-1], "/"), strings.Join(arr[l-1:l], "")
}
