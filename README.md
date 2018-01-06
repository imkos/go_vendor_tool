# go get vendor by Golang

> 使用说明

```
Usage of go_vendor_tool.exe, Version: 1.0.2
go vendor get! by K.o.s[vbz276@gmail.com]!
  -dir string
        Set Home RootPath (default ".")
  -env
        Show GOPATH
  -usegoenv
        Use GOPATH; Otherwise use project[RootPath] vendor folder
```



> 依赖工具
```
1. 此工具是基于govendor(https://github.com/kardianos/govendor)生成的vendor.json进行go get.
2. go get 需系统已安装git.
```



> 生成vendor.json

```shell
govendor init

govendor list

govendor add +external
```

> 部署

1. 项目或产品代码发布时包含vendor目录，并提交vendor下的vendor.json的文件
2. 如使用--usegoenv 参数表示使用系统的GOPATH, 会把go get的库放在GOPATH[0]的src目录下

