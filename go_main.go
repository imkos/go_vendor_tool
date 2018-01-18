package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	Major_Ver = "1.0.2"
	//
	VENDOR_DIR       = "vendor"
	VERDOR_DESC_FILE = VENDOR_DIR + ".json"
	sep_path         = string(filepath.Separator)
)

var (
	s_gopath string
	wg       *sync.WaitGroup
	//vars
	s_home_rootpath   *string
	bo_show_gopath    *bool
	bo_use_sys_gopath *bool
	mp_cache          = &sync.Map{}
)

type st_verdor_package struct {
	Revisiontime string `json:"revisionTime"`
	Path         string `json:"path"`
	Checksumsha1 string `json:"checksumSHA1"`
	Revision     string `json:"revision"`
}

type st_verdor_info struct {
	Comment  string               `json:"comment"`
	Ignore   string               `json:"ignore"`
	Rootpath string               `json:"rootPath"`
	Package  []*st_verdor_package `json:"package"`
}

//判断文件或文件夹是否存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func execCommand_with_ospipe(commandName string, params []string, Dir_env string) bool {
	cmd := exec.Command(commandName, params...)
	cmd.Dir = Dir_env
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Stdout = os.Stdout // 重定向标准输出
	cmd.Stderr = os.Stderr // 重定向标准输出
	done := make(chan error)
	go func() {
		fmt.Println("Dir:", Dir_env)
		done <- cmd.Run()
	}()
	select {
	case err := <-done:
		if err != nil {
			close(done)
			fmt.Println("Wait error:", err)
			return false
		}
		fmt.Println(buf.String())
		return true
	}
}

func execCommand(commandName string, params []string, Dir_env string) bool {
	cmd := exec.Command(commandName, params...)
	cmd.Dir = Dir_env
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return false
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Start error:", err, Dir_env, commandName, params)
		return false
	}
	buf := &bytes.Buffer{}
	done := make(chan error)
	go func() {
		if _, err := buf.ReadFrom(stdout); err != nil {
			panic("buf.Read(stdout) error: " + err.Error())
		}
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
			close(done)
			fmt.Println("Wait error:", err, Dir_env, commandName, params)
			return false
		}
		fmt.Println("Dir:", Dir_env)
		fmt.Println(buf.String())
		return true
	}
}

func go_get_pkg(pkg *st_verdor_package, ch <-chan struct{}) {
	defer func() {
		wg.Done()
		<-ch
	}()
	if pkg != nil {
		//处理常用golang.org/x 系列
		if strings.Contains(pkg.Path, "golang.org/x") {
			a_str := strings.Split(pkg.Path, "/")
			var new_dir string
			var pkg_name string
			for i, v := range a_str {
				if v == "x" || a_str[i+2] != "" {
					pkg_name = a_str[i+2]
					new_dir = filepath.Join(s_gopath, strings.Join(a_str[:i+2], "/"))
					break
				}
			}
			if _, ok := mp_cache.Load("golang.org/x/" + pkg_name); ok {
				return
			} else {
				mp_cache.Store("golang.org/x/"+pkg_name, 1)
			}
			/*
				$ git clone $URL
				$ cd $PROJECT_NAME
				$ git reset --hard $SHA1
			*/
			if !Exist(new_dir) {
				os.MkdirAll(new_dir, 0755)
				mirror_path := "github.com/golang/" + pkg_name
				execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
				execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
			} else {
				pkg_dir := filepath.Join(new_dir, pkg_name)
				if Exist(pkg_dir) {
					if Exist(filepath.Join(pkg_dir, ".git")) {
						execCommand("git", []string{"pull"}, filepath.Join(new_dir, pkg_name))
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					} else {
						os.RemoveAll(pkg_dir)
						mirror_path := "github.com/golang/" + pkg_name
						execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					}
				} else {
					mirror_path := "github.com/golang/" + pkg_name
					execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
					execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
				}
			}
			//处理google.golang.org 域名下的grpc与genproto(TODO:暂时不处理)
		} else if strings.Contains(pkg.Path, "google.golang.org") {

			//处理google/ 域名下的一些库
		} else if strings.Contains(pkg.Path, "google/") {
			var new_dir string
			var pkg_name string
			var pkg_url string
			if strings.Count(pkg.Path, "/") > 1 {
				pkg_fields := strings.Split(pkg.Path, "/")
				pkg_url = strings.Join(pkg_fields[:2], "/")
				if _, ok := mp_cache.Load(pkg_url); ok {
					return
				} else {
					mp_cache.Store(pkg_url, 1)
					ln := strings.LastIndex(pkg_url, "/")
					new_dir = filepath.Join(s_gopath, pkg_url[:ln])
					pkg_name = pkg_url[ln+1:]
				}
			} else {
				//由于govendor工具生成的path都是已排序的,其实此处的判断可以省略
				if _, ok := mp_cache.Load(pkg.Path); ok {
					return
				} else {
					mp_cache.Store(pkg.Path, 1)
					pkg_url = pkg.Path
					ln := strings.LastIndex(pkg.Path, "/")
					new_dir = filepath.Join(s_gopath, pkg.Path[:ln])
					pkg_name = pkg.Path[ln+1:]
				}
			}
			if !Exist(new_dir) {
				os.MkdirAll(new_dir, 0755)
				mirror_path := "github.com/" + pkg_url
				execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
				execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
			} else {
				pkg_dir := filepath.Join(new_dir, pkg_name)
				if Exist(pkg_dir) {
					if Exist(filepath.Join(pkg_dir, ".git")) {
						execCommand("git", []string{"pull"}, filepath.Join(new_dir, pkg_name))
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					} else {
						os.RemoveAll(pkg_dir)
						mirror_path := "github.com/" + pkg_url
						execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					}
				} else {
					mirror_path := "github.com/" + pkg_url
					execCommand("git", []string{"clone", "https://" + mirror_path + ".git"}, new_dir)
					execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
				}
			}
		} else {
			var new_dir string
			var pkg_name string
			var pkg_url string
			if strings.Count(pkg.Path, "/") > 2 {
				pkg_fields := strings.Split(pkg.Path, "/")
				pkg_url = strings.Join(pkg_fields[:3], "/")
				if _, ok := mp_cache.Load(pkg_url); ok {
					return
				} else {
					mp_cache.Store(pkg_url, 1)
					ln := strings.LastIndex(pkg_url, "/")
					new_dir = filepath.Join(s_gopath, pkg_url[:ln])
					pkg_name = pkg_url[ln+1:]
				}
			} else {
				//由于govendor工具生成的path都是已排序的,其实此处的判断可以省略
				if _, ok := mp_cache.Load(pkg.Path); ok {
					return
				} else {
					mp_cache.Store(pkg.Path, 1)
					pkg_url = pkg.Path
					ln := strings.LastIndex(pkg.Path, "/")
					new_dir = filepath.Join(s_gopath, pkg.Path[:ln])
					pkg_name = pkg.Path[ln+1:]
				}
			}
			if !Exist(new_dir) {
				os.MkdirAll(new_dir, 0755)
				execCommand("git", []string{"clone", "https://" + pkg_url + ".git"}, new_dir)
				execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
			} else {
				pkg_dir := filepath.Join(new_dir, pkg_name)
				if Exist(pkg_dir) {
					if Exist(filepath.Join(pkg_dir, ".git")) {
						execCommand("git", []string{"pull"}, filepath.Join(new_dir, pkg_name))
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					} else {
						os.RemoveAll(pkg_dir)
						execCommand("git", []string{"clone", "https://" + pkg_url + ".git"}, new_dir)
						execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
					}
				} else {
					execCommand("git", []string{"clone", "https://" + pkg_url + ".git"}, new_dir)
					execCommand("git", []string{"reset", "--hard", pkg.Revision}, filepath.Join(new_dir, pkg_name))
				}
			}
		}
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s, Version: %s\n", os.Args[0], Major_Ver)
		fmt.Println("go vendor get! by K.o.s[vbz276@gmail.com]!")
		flag.PrintDefaults()
	}
	s_home_rootpath = flag.String("dir", filepath.Dir(os.Args[0]), "Set Home RootPath")
	bo_show_gopath = flag.Bool("env", false, "Show GOPATH")
	bo_use_sys_gopath = flag.Bool("usegoenv", false, "Use GOPATH; Otherwise use project[RootPath] vendor folder")
}

func main() {
	flag.Parse()
	//
	if *bo_show_gopath {
		fmt.Println("go env|[GOPATH]:", os.Getenv("GOPATH"))
		return
	}
	s_path := *s_home_rootpath
	//
	s_vendor_dir := filepath.Join(s_path, VENDOR_DIR)
	if !Exist(s_vendor_dir) {
		fmt.Printf("[%s] is not Exist\n", s_vendor_dir)
		return
	}
	s_verdor_desc_file := s_vendor_dir + sep_path + VERDOR_DESC_FILE
	if !Exist(s_verdor_desc_file) {
		fmt.Printf("[%s] is not Exist\n", s_verdor_desc_file)
		return
	}
	b_json, e1 := ioutil.ReadFile(s_verdor_desc_file)
	if e1 != nil {
		fmt.Println("verdor_desc_file read error:", e1)
		return
	}
	var vi st_verdor_info
	if e2 := json.Unmarshal(b_json, &vi); e2 != nil {
		fmt.Println("verdor_desc_file read error:", e2)
		return
	}
	if *bo_use_sys_gopath {
		//%path%变量中的分隔符
		g_gopath := filepath.SplitList(os.Getenv("GOPATH"))
		if len(g_gopath) > 1 && len(g_gopath[0]) > 0 {
			s_gopath = g_gopath[0] + sep_path + "src"
		}
		if s_gopath == "" {
			fmt.Println("GOPATH IS invalid!")
			return
		}
	} else {
		s_gopath = s_vendor_dir
	}
	if !Exist(s_gopath) {
		os.MkdirAll(s_gopath, 0755)
	}
	wg = new(sync.WaitGroup)
	fmt.Println("go get start ....", s_gopath)
	ch_worker := make(chan struct{}, 5)
	if len(vi.Package) > 0 {
		for _, pkg := range vi.Package {
			ch_worker <- struct{}{}
			wg.Add(1)
			go go_get_pkg(pkg, ch_worker)
		}
	}
	//等待所有操作完成
	wg.Wait()
	fmt.Println("go get finish!")
}
