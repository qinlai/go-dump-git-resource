package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var gGitDir, gTargetDir string
var gProject, gBaseTag, gEndTag string
var gIsPull bool
var gProjectDir string

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
}

func main() {
	flag.Parse()
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()
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

	if gIsPull {
		gitPull(gProjectDir)
	}

	if len(gEndTag) > 0 {
		doDiff(
			gProjectDir,
			fmt.Sprintf("%s/%s", gTargetDir, FILE_DIR_NAME),
			fmt.Sprintf("%s/%s..%s", diffDir, gBaseTag, gEndTag),
			getGitInfo(gProjectDir, []string{"diff-tree", "-r", "-t", fmt.Sprintf("%s..%s", gBaseTag, gEndTag)}))
	} else {
		doBase(
			gProjectDir,
			fmt.Sprintf("%s/%s", gTargetDir, TREE_DIR_NAME),
			fmt.Sprintf("%s/%s", indexDir, gBaseTag),
			getGitInfo(gProjectDir, []string{"ls-tree", "-r", "-t", gBaseTag}))
	}

	println("completed")
}

func doBase(fromDir string, toDir string, indexFile string, lines []string) {
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
		info := reg.FindAll([]byte(line), 4)
		if len(info) < 4 {
			continue
		}

		t := string(info[1])
		sha := string(info[2])
		fullName := string(info[3])

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
			if e = exec.Command("cp", "-u", from, to).Run(); e != nil {
				panic(e)
			}

			hsd1, _ := hex.DecodeString(fmt.Sprintf("%x", sha1.Sum([]byte(path)))[:10])
			hsd2, _ := hex.DecodeString(dirs[path])

			f.Write(hsd1)
			f.Write(hsd2)
		}
	}
}

func doDiff(fromDir string, toDir string, diffFile string, lines []string) {
	if !isFileExists(toDir) {
		os.Mkdir(toDir, os.ModePerm)
	}

	f, e := os.Create(diffFile)
	if e != nil {
		panic(e)
	}

	for _, line := range lines {
		arr := strings.Split(line, " ")
		if len(arr) != 5 {
			continue
		}

		path := arr[3]
		fileArr := strings.Split(arr[4], "\t")
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

		if e = exec.Command("cp", "-u", fromFile, to).Run(); e != nil {
			panic(e)
		}

		hsd1, _ := hex.DecodeString(fmt.Sprintf("%x", sha1.Sum([]byte(fullFile)))[:10])
		hsd2, _ := hex.DecodeString(path)

		f.Write(hsd1)
		f.Write(hsd2)
	}
}

func gitPull(gitDir string) {
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

func getGitInfo(cmdDir string, args []string) []string {
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
	if _, e := os.Open(file); e != nil {
		return false
	}

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
