package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_str(t *testing.T) {
	s := "github.com/golang/protobuf/proto"
	ln := strings.LastIndex(s, "/")
	fmt.Println(ln, s[:ln], s[ln+1:])
	s1 := "golang.org/x/sys/unix"
	a_str := strings.Split(s1, "/")
	fmt.Println(strings.Join(a_str[:3], "/"))
	s_gopath := filepath.Dir(os.Args[0])
	var new_dir string
	for i, v := range a_str {
		if v == "x" || a_str[i+2] != "" {
			fmt.Println(a_str[:i+3])
			new_dir = filepath.Join(s_gopath, strings.Join(a_str[:i+3], "/"))
			break
		}
	}
	fmt.Println(new_dir)
}

func Test_exec_cmd(t *testing.T) {
	execCommand("git", []string{"clone", "https://github.com/imkos/hashsum.git"}, filepath.Dir(os.Args[0]))
}

func Test_some(t *testing.T) {
	s_gopath = filepath.Join(`E:\goWork\src\go_vendor_tool`, VENDOR_DIR)
	ch_worker := make(chan struct{}, 5)
	pkg := &st_verdor_package{
		Path:     "github.com/boombuler/barcode",
		Revision: "3cfea5ab600ae37946be2b763b8ec2c1cf2d272d",
	}
	go_get_pkg(pkg, ch_worker)
}
