package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func handle_request() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		request := *ctx.Request
		log.Printf("-------------- Request Headers --------------")
		log.Printf("Address: %s", request.RemoteAddr)
		log.Printf("Method: %s", request.Method)
		log.Printf("Path: %s", request.URL.String())
		if request.ContentLength > 0 {
			log.Printf("ContentLength: %d Byte\n", request.ContentLength)
		}
		for k, v := range request.Header {
			log.Printf("\t%s: %s", k, v)
		}

		log.Printf("-------------- Response Headers -------------")
		log.Printf("Status: %d %s", ctx.Writer.Status(), http.StatusText(ctx.Writer.Status()))
		log.Printf("\tServer: xserver")
		for k, v := range ctx.Writer.Header() {
			log.Printf("\t%s: %s", k, v)
		}
		log.Printf("Duration: %v", time.Since(start))
		log.Printf("---------------- Request End ----------------")
		log.Println()
	}
}

func init() {
	log.SetFlags(log.LstdFlags)
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	port := flag.Int("p", 22345, "http server port")
	dir := flag.String("d", ".", "server root directory")
	help := flag.Bool("h", false, "Show help message")
	flag.Parse()

	if *help {
		log.Println("Usage: xserver [OPTIONS]")
		log.Println()
		log.Println("Options:")
		log.Println("  -p int      Specify the port to run the HTTP server (default: 22345)")
		log.Println("  -d string   Specify the server directory for the HTTP server (default: ./ )")
		log.Println("  -h          Show help message")
		log.Println()
		log.Println("example: xserver -p 22345 -d /")
		os.Exit(0)
	}

	absolute_dir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("directory %s（error：%v）", *dir, err)
	}

	file_info, err := os.Stat(absolute_dir)
	if os.IsNotExist(err) {
		log.Fatalf("directory %s not exists", absolute_dir)
	}
	if !file_info.IsDir() {
		log.Fatalf("directory %s is invalid", absolute_dir)
	}

	engine := gin.Default()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(handle_request())
	engine.Use(static.Serve("/", static.LocalFile(absolute_dir, true)))
	log.Printf("X server start:\n")
	log.Printf("dir: %s\n", absolute_dir)
	log.Printf("port: %d\n", *port)
	log.Println("host: ")
	log.Println("  http://localhost:" + strconv.Itoa(*port))
	log.Println("  http://127.0.0.1:" + strconv.Itoa(*port))
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("error: %v", err)
	}
}
