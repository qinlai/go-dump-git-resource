package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var gBranch, gGitDir, gTargetDir string
var gProject, gBaseTag, gEndTag string
var gIsPull, gIsOpenLog bool
var gProjectDir string
var gFilePathReg *regexp.Regexp
var gForceCopy bool

const (
	TREE                string = "tree"
	BLOB                string = "blob"
	TREE_DIR_NAME       string = "tree"
	FILE_DIR_NAME       string = "file"
	INDEX_FILE_DIR_NAME string = "index"
	DIFF_FILE_DIR_NAME  string = "diff"
)

func init() {
	flag.StringVar(&gGitDir, "git_dir", "", "git dir(request)")
	flag.StringVar(&gTargetDir, "target_dir", "", "target dir(request)")
	flag.StringVar(&gProject, "project", "", "project name(request)")
	flag.StringVar(&gBaseTag, "base_tag", "", "base git tag name(request)")
	flag.StringVar(&gEndTag, "end_tag", "", "end git tag name")
	flag.BoolVar(&gIsPull, "is_pull", false, "is need pull before dump?")
	flag.StringVar(&gBranch, "branch", "master", "git branch")
	flag.BoolVar(&gIsOpenLog, "is_openlog", true, "print log?")
	flag.BoolVar(&gForceCopy, "force_copy", false, "force copy file")
}

func main() {
	flag.Parse()
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	var err error
	gFilePathReg, err = regexp.Compile(" ")
	if err != nil {
		panic(err.Error())
	}

	do()
}

func do() {
	gProjectDir = fmt.Sprintf("%s/%s", gGitDir, gProject)
	gTargetDir = fmt.Sprintf("%s/%s", gTargetDir, gProject)
	indexDir := fmt.Sprintf("%s/%s", gTargetDir, INDEX_FILE_DIR_NAME)
	diffDir := fmt.Sprintf("%s/%s", gTargetDir, DIFF_FILE_DIR_NAME)
	check()

	if !isFileExists(gTargetDir) {
		os.Mkdir(gTargetDir, os.ModePerm)
	}

	if !isFileExists(indexDir) {
		os.Mkdir(indexDir, os.ModePerm)
	}

	if !isFileExists(diffDir) {
		os.Mkdir(diffDir, os.ModePerm)
	}

	gitCheckout(gProjectDir, gBranch)
	if gIsPull {
		gitPull(gProjectDir)
	}
	gitCheckout(gProjectDir, gBaseTag)

	doBase(
		gProjectDir,
		fmt.Sprintf("%s/%s", gTargetDir, TREE_DIR_NAME),
		fmt.Sprintf("%s/%s", indexDir, gBaseTag),
		getGitInfo(gProjectDir, []string{"ls-tree", "-r", "-t", gBaseTag}))

	if len(gEndTag) > 0 {
		gitCheckout(gProjectDir, gEndTag)
		doDiff(
			gProjectDir,
			fmt.Sprintf("%s/%s", gTargetDir, FILE_DIR_NAME),
			fmt.Sprintf("%s/%s..%s", diffDir, gBaseTag, gEndTag),
			getGitInfo(gProjectDir, []string{"diff-tree", "-r", "-t", fmt.Sprintf("%s..%s", gBaseTag, gEndTag)}))
	}

	println("completed")
}

func doBase(fromDir string, toDir string, indexFile string, lines []string) {
	printLog("dump base")
	if !isFileExists(toDir) {
		os.Mkdir(toDir, os.ModePerm)
	}

	dirs := make(map[string]string)
	reg, e := regexp.Compile(`(\S+)`)

	if e != nil {
		panic(e)
	}

	f, e := os.Create(indexFile)
	if e != nil {
		panic(e)
	}

	for _, line := range lines {
		info := reg.FindAll([]byte(line), math.MaxInt32)
		if len(info) < 4 {
			continue
		}

		t := string(info[1])
		sha := string(info[2])
		fullName := string(bytes.Join(info[3:], []byte(" ")))

		if t == TREE {
			dirs[fullName] = sha
			treeDir := fmt.Sprintf("%s/%s", toDir, dirs[fullName])
			if !isFileExists(treeDir) {
				os.Mkdir(treeDir, os.ModePerm)
			}
		} else if t == BLOB {
			path, file := getFileInfo(fullName)
			from := fmt.Sprintf("%s/%s", fromDir, fullName)
			to := fmt.Sprintf("%s/%s/%s", toDir, dirs[path], file)
			cp(from, to)

			hsd1, _ := hex.DecodeString(fmt.Sprintf("%x", sha1.Sum([]byte(path)))[:10])
			hsd2, _ := hex.DecodeString(dirs[path])

			if len(hsd2) == 0 {
				continue
			}
			f.Write(hsd1)
			f.Write(hsd2)
		}
	}
	printLog("dump base completed")
}

func doDiff(fromDir string, toDir string, diffFile string, lines []string) {
	printLog("dump diff")

	if !isFileExists(toDir) {
		os.Mkdir(toDir, os.ModePerm)
	}

	f, e := os.Create(diffFile)
	if e != nil {
		panic(e)
	}

	for _, line := range lines {
		arr := strings.Split(line, " ")
		if len(arr) < 5 {
			continue
		}

		path := arr[3]
		fileArr := strings.Split(strings.Join(arr[4:], " "), "\t")
		fullFile := fileArr[len(fileArr)-1]
		_, file := getFileInfo(fullFile)

		fromFile := fmt.Sprintf("%s/%s", fromDir, fullFile)
		if isDir(fromFile) {
			continue
		}

		diffDir := fmt.Sprintf("%s/%s", toDir, path)
		if !isFileExists(diffDir) {
			os.Mkdir(diffDir, os.ModePerm)
		}

		to := fmt.Sprintf("%s/%s", diffDir, file)

		cp(fromFile, to)

		hsd1, _ := hex.DecodeString(fmt.Sprintf("%x", sha1.Sum([]byte(fullFile)))[:10])
		hsd2, _ := hex.DecodeString(path)

		if len(hsd2) == 0 {
			continue
		}

		f.Write(hsd1)
		f.Write(hsd2)
	}

	printLog("dump diff completed.")
}

func gitPull(gitDir string) {
	printLog("git pull")
	shell, e := exec.LookPath("git")
	if e != nil {
		panic(e)
	}

	cmd := new(exec.Cmd)
	cmd.Dir = gitDir
	cmd.Path = shell
	cmd.Args = append([]string{"git", "pull"})

	if e := cmd.Run(); e != nil {
		panic(e)
	}
}

func gitCheckout(gitDir, tag string) {
	printLog("git checkout %s", tag)
	shell, e := exec.LookPath("git")
	if e != nil {
		panic(e)
	}

	cmd := new(exec.Cmd)
	cmd.Dir = gitDir
	cmd.Path = shell
	cmd.Args = append([]string{"git", "checkout", tag})

	if e := cmd.Run(); e != nil {
		panic(e)
	}
	printLog("git checkout completed")
}

func getGitInfo(cmdDir string, args []string) []string {
	printLog("get git info %#v", args)
	shell, e := exec.LookPath("git")
	if e != nil {
		panic(e)
	}

	cmd := new(exec.Cmd)
	cmd.Dir = gProjectDir
	cmd.Path = shell
	cmd.Args = append([]string{"git"}, args...)

	r, e := cmd.Output()
	if e != nil {
		panic(e)
	}

	printLog("get git info completed")
	return strings.Split(string(r), "\n")
}

func check() {
	if len(gGitDir) < 1 {
		panic(fmt.Errorf("%s", "git_dir can't be null"))
	}

	if len(gTargetDir) < 1 {
		panic(fmt.Errorf("%s", "cdn_dir can't be null"))
	}

	if len(gProject) < 1 {
		panic(fmt.Errorf("%s", "project can't be null"))
	}

	if len(gBaseTag) < 1 {
		panic(fmt.Errorf("%s", "base_tag can't be null"))
	}

	if !isFileExists(gGitDir) {
		panic(fmt.Errorf("%s", "git_dir file not exists"))
	}

	if !isFileExists(gProjectDir) {
		panic(fmt.Errorf("%s", "project file not exists"))
	}
}

func isFileExists(file string) bool {
	f, e := os.Open(file)
	if e != nil {
		return false
	}
	f.Close()
	return true
}

func isDir(dir string) bool {
	f, e := os.Open(dir)
	if e != nil {
		return false
	}

	fi, e := f.Stat()

	if e != nil {
		panic(e)
	}

	f.Close()
	return fi.IsDir()
}

func getFileInfo(file string) (dir, fileName string) {
	arr := strings.Split(file, "/")
	l := len(arr)
	return strings.Join(arr[0:l-1], "/"), strings.Join(arr[l-1:l], "")
}

func cp(from, to string) {
	_, err := os.Stat(to)
	if gForceCopy || (err != nil && !os.IsExist(err)) {
		if e := exec.Command("cp", "-u", from, to).Run(); e != nil {
			printLog("cp error %s=>%s=>%s", from, to, e.Error())
			panic(e.Error())
		}
		printLog("cp %s=>%s", from, to)
	}
}

func printLog(format string, a ...interface{}) {
	if gIsOpenLog {
		log.Printf(format+"\n", a...)
	}
}
