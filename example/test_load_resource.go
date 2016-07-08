/*
示例读取项目project1的文件(资源抽取目录为dump/project1)
*/
package main

import "github.com/qinlai/go-dump-git-resource/load_resource"

const (
	//http|path
	uri = "../example/dump/project1/"

	baseTag = "201601"

	endTag = "201602"
)

func main() {
	f1, f2 := "test/t.txt", "test2/t2.txt"

	loader, e := loadresource.NewLoader(uri, baseTag, endTag)
	if e != nil {
		panic(e)
	}

	loader.IsDebug = true
	d1, e1 := loader.LoadData(f1)
	if e1 != nil {
		panic(e1)
	}

	d2, e2 := loader.LoadData(f2)
	if e2 != nil {
		panic(e2)
	}

	println("=============read file==================")
	println(f1, ":", string(d1))
	println("========================================")
	println(f2, ":", string(d2))
	println("=============end read file==============")
}
