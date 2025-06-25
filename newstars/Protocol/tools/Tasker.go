package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	reGen = false
	build = false
)

/** 获取程序根目录 */
func GetRootPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file + "/../..")

	return strings.Replace(path, "\\", "/", -1)
}

/** 判断目录或文件是否存在 */
func CheckFile(path string) int {
	f, err := os.Stat(path)
	if err != nil {
		return -1
	}
	switch mode := f.Mode(); {
	case mode.IsDir():
		return 1 // 目录
	case mode.IsRegular():
		return 2 // 普通文件
	case mode&os.ModeSymlink != 0:
		return 3 // 软连接
	}

	return 0 // 未知
}

/** 执行命令 */
func Exec(buffer *bytes.Buffer, command string, params ...string) bool {
	cmd := exec.Command(command, params...)
	//fmt.Println(cmd.Args)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	cmd.Start()
	reader := bufio.NewReader(stdout)
	var line []byte
	for {
		line, _, err = reader.ReadLine()
		if err != nil || io.EOF == err {
			if err != io.EOF {
				return false
			}
			break
		}
		if buffer != nil {
			buffer.Write(line)
		} else {
			fmt.Println(string(line))
		}
	}

	cmd.Wait()
	if stderr.Len() > 0 {
		fmt.Println("Exec Out:", stderr.String())
		return false
	}
	return true
}

/** 初始化 */
func Init() {
	for _, v := range os.Args[1:] {
		switch strings.ToLower(v) {
		case "regen":
			reGen = true
		case "build":
			build = true
		}
	}
}

/** 复制文件 */
func CopyFile(source, dest string) bool {
	sf, err := os.Open(source)
	if err != nil {
		return false
	}
	defer sf.Close()
	sd, err := ioutil.ReadAll(sf)

	df, err := os.Create(dest)
	if err != nil {
		return false
	}
	defer df.Close()
	df.Write(sd)

	return true
}

/** 生成目标Proto */
func CreatorProto(path string, protofile string) bool {
	pf, err := os.Create(protofile)
	if err != nil {
		return false
	}

	defer pf.Close()
	pf.WriteString("syntax = \"proto3\";\npackage plr;\n\noption go_package=\"./\";\n\n")

	filepath.Walk(path, func(filename string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		if fi.Size() > 0 && strings.HasSuffix(strings.ToLower(fi.Name()), "proto") {
			fi, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			fd, err := ioutil.ReadAll(fi)
			fi.Close()
			pf.Write(fd)
		}
		return nil
	})

	return true
}

/** 替换资源 */
func sub(file string, old string, need string) bool {
	pbf, err := os.Open(file)
	if err != nil {
		return false
	}
	pbd, err := ioutil.ReadAll(pbf)
	pbf.Close()

	newf, err := os.Create(file)
	if err != nil {
		return false
	}
	newf.Write(bytes.Replace(pbd, []byte(old), []byte(need), 1))
	newf.Close()

	return true
}

/** 编译程序 */
func Build() bool {
	root := GetRootPath()
	toolPath := path.Join(root, "tools")
	protoc := path.Join(toolPath, "bin", "protoc")

	var fmod string
	if runtime.GOOS == "windows" {
		fmod = "%s;%s"
	} else {
		if runtime.GOOS == "darwin" {
			protoc = protoc + "-mac"
		}
		fmod = "%s:%s"
	}

	os.Chdir(root)
	os.Setenv("PATH", fmt.Sprintf(fmod, os.Getenv("PATH"), root))

	if CheckFile("./Build") > 0 {
		os.RemoveAll(path.Join(root, "Build"))
	}

	os.MkdirAll("Build/tmp", 0744)
	os.MkdirAll("Build/js", 0744)
	os.MkdirAll("Build/go", 0744)

	if !CreatorProto("src", "Build/tmp/Protos.proto") {
		return false
	}

	// if !Exec(nil, protoc, "-IBuild/tmp", "--js_out=import_style=commonjs,binary:Build/tmp", "Build/tmp/Protos.proto") {
	// 	return false
	// }

	// os.Chdir("Build/tmp")
	// if !Exec(nil, "npm", "init", "-y") {
	// 	return false
	// }

	// Exec(nil, "npm", "install", "google-protobuf")
	// CopyFile("../../tools/temp/webpack.config.js", "webpack.config.js")
	// CopyFile("../../tools/temp/main.js", "main.js")
	// CopyFile("../../tools/temp/google-protobuf.js", "google-protobuf.js")

	// sub("Protos_pb.js", "var goog = jspb;", "var goog = jspb;\ngoog.DEBUG=true;\nCOMPILED=false;")

	// webpack := "./node_modules/.bin/webpack"
	// if CheckFile(webpack) < 0 {
	// 	Exec(nil, "npm", "install", "webpack", "--save-dev")
	// 	Exec(nil, "npm", "install", "webpack-cli", "--save-dev")
	// }
	// if !Exec(nil, webpack, "--mode", "production") {
	// 	fmt.Println("执行 WebPack 异常！")
	// 	return false
	// }
	os.Chdir(root)

	// 生成GO版本协议
	pggo := "--plugin=protoc-gen-go=tools/protoc-gen-go/bin/protoc-gen-go"
	if runtime.GOOS == "windows" {
		pggo += ".exe"
	}
	if !Exec(nil, protoc, pggo, "-IBuild/tmp", "--go_out=Build/go", "Build/tmp/Protos.proto") {
		return false
	}

	// 加入版本好
	TIME := time.Now().UnixNano() / 1000000000
	fmt.Println("协议版本:", TIME)
	sub("Build/js/ProtoLib.js", "pslib.VERSION = 0;", fmt.Sprintf("pslib.VERSION = %d;", TIME))
	sub("Build/go/Protos.pb.go", "const _ = proto.ProtoPackageIsVersion2", fmt.Sprintf("var VERSION = %d\nconst _ = proto.ProtoPackageIsVersion2", TIME))
	sub("Build/go/Protos.pb.go", "package __", "package plr")

	return true
}

/** 获取/更新依赖库 */
func ReGen() bool {
	root := GetRootPath()
	genPath := path.Join(root, "tools", "protoc-gen-go")
	os.Setenv("GOPATH", genPath)
	if CheckFile(genPath) <= 0 {
		os.MkdirAll(genPath, 0744)
		if !Exec(nil, "go", "get", "-d", "github.com/golang/protobuf/protoc-gen-go") {
			log.Println("拉取 Protoc-gen-go 失败")
			return false
		}
	}

	defer os.Remove(path.Join(genPath, "pkg"))
	//os.Setenv("GOBIN", path.Join(root, "tools", "bin"))
	if !Exec(nil, "go", "install", "github.com/golang/protobuf/protoc-gen-go") {
		log.Println("编译 protoc-gen-go 失败")
		return false
	}

	return true
}

func main() {
	Init()
	if reGen {
		log.Println("更新 Protoc-gen-go")
		if !ReGen() {
			return
		}
	}

	if build {
		log.Println("编译程序")
		log.Println(Build())
	}
}
