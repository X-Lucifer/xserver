### X Server

> 使用`Gin`框架实现的高性能静态文件服务器

#### 编译资源文件
- ##### 安装依赖
> go install github.com/tc-hib/go-winres@latest

- ##### 初始化
> go-winres init

- ##### 构建资源
> go-winres make

#### 打包构建
> go build

#### 使用方式
    Usage: xserver [OPTIONS]
    
    Options:
        -p int      Specify the port to run the HTTP server (default: 22345)
        -d string   Specify the server directory for the HTTP server (default: ./ )
        -h          Show help message

    example: .\xserver -p 12345 -d dist