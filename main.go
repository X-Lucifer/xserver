package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func HandlerContext() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		request_id := uuid.New().String()
		request := *ctx.Request
		logger := log.With().Str("RequestId", request_id).Logger()
		ctx.Header("X-Request-Id", request_id)
		ctx.Header("Server", "xserver")
		request_log_info := logger.Info()
		request_log_info.Str("Client", ctx.ClientIP()).Str("Remote", request.RemoteAddr).Str("Method", request.Method).Str("Path", request.URL.String())
		if request.ContentLength > 0 {
			request_log_info.Int64("ContentLength", request.ContentLength)
		}
		for k, v := range request.Header {
			request_log_info.Strs(k, v)
		}
		request_log_info.Msgf("Request Headers\n")
		ctx.Next()
		response_log_info := logger.Info()
		response_log_info.Int("StatusCode", ctx.Writer.Status()).Str("Status", http.StatusText(ctx.Writer.Status()))
		for k, v := range ctx.Writer.Header() {
			response_log_info.Strs(k, v)
		}
		response_log_info.Str("Duration", time.Since(start).String()).Msgf("Response Headers\n")
	}
}

func GetLocalUrls(port int) []string {
	var urls []string
	urls = append(urls, fmt.Sprintf("http://127.0.0.1:%d", port))
	urls = append(urls, fmt.Sprintf("http://localhost:%d", port))
	interfaces, err := net.Interfaces()
	if err != nil {
		return urls
	}

	for _, inet := range interfaces {
		addrs, err := inet.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}

			if ip.IsLinkLocalUnicast() {
				continue
			}

			if ip.To4() != nil {
				urls = append(urls, fmt.Sprintf("http://%s:%d", ip.String(), port))
			} else {
				urls = append(urls, fmt.Sprintf("http://[%s]:%d", ip.String(), port))
			}
		}
	}

	seen := make(map[string]bool)
	var unique []string
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			unique = append(unique, u)
		}
	}
	return unique
}

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log_output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		NoColor:    false,
		PartsOrder: []string{"time", "level", "message"},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s\n", i)
		},
	}
	log.Logger = log.Output(log_output)
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	port_str := flag.String("p", "22345", "http server port")
	dir := flag.String("d", ".", "server root directory")
	help := flag.Bool("h", false, "Show help message")
	flag.Parse()
	if *help {
		fmt.Println("Usage: xserver [OPTIONS]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -p int      Specify the port to run the HTTP server (default: 22345)")
		fmt.Println("  -d string   Specify the server directory for the HTTP server (default: ./ )")
		fmt.Println("  -h          Show help message")
		fmt.Println()
		fmt.Println("example: xserver -p 22345 -d /")
		os.Exit(0)
	}

	port := 22345
	if xport, err := strconv.Atoi(*port_str); err == nil && xport >= 1000 && xport <= 65535 {
		port = xport
	} else {
		log.Warn().Str("invalid_port", *port_str).Msg("port is invalid, using default 22345")
	}
	absolute_dir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal().Err(err).Msgf("directory %s (error: %v)\n", *dir, err)
	}

	file_info, err := os.Stat(absolute_dir)
	if os.IsNotExist(err) {
		log.Fatal().Err(err).Msgf("directory %s not exists\n", absolute_dir)
	}
	if !file_info.IsDir() {
		log.Fatal().Err(err).Msgf("directory %s is invalid\n", absolute_dir)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	_ = engine.SetTrustedProxies(nil)
	engine.Use(HandlerContext())
	engine.StaticFS("/", gin.Dir(absolute_dir, false))
	log.Info().Msg("X server start")
	log.Info().Str("dir", absolute_dir).Int("port", port).Strs("urls", GetLocalUrls(port)).Msgf("Server:\n")
	addr := fmt.Sprintf(":%d", port)
	if err := engine.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Exception")
	}
}
