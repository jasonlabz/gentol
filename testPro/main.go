package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jasonlabz/potato/configx"
	"github.com/jasonlabz/potato/ginmetrics"

	"testPro/bootstrap"
	"testPro/server/routers"
)

// @title		    TODO: ***********服务
// @version		    1.0
// @description	    TODO: 旨在***********
// @host			TODO: localhost:port
// @contact.name	TODO: your name
// @contact.url	    TODO: http://www.*****.io/support
// @contact.email	TODO: mail_name@qq.com
// @BasePath		TODO: /base_path
func main() {
	// context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// bootstrap init
	bootstrap.MustInit(ctx)

	// gin mode
	serverMode := gin.ReleaseMode
	serverConfig := configx.GetConfig()
	if serverConfig.Debug {
		serverMode = gin.DebugMode
	}
	gin.SetMode(serverMode)

	r := routers.InitApiRouter()

	appConf := serverConfig.Application
	if appConf.Prom.Enable {
		// get global Monitor object
		m := ginmetrics.GetMonitor()

		// +optional set metric path, default /debug/metrics
		m.SetMetricPath(appConf.Prom.Path)
		// +optional set slow time, default 5s
		m.SetSlowTime(10)
		// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
		// used to p95, p99
		m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

		// set middleware for gin
		m.Use(r)
	}

	if appConf.PProf.Enable {
		r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", appConf.PProf.Port), nil); err != nil {
				log.Fatalf("pprof server failed: %v", err)
			}
		}()
	}

	if appConf.FileServer {
		go func() {
			fileServer(appConf.Port + 1)
		}()
	}

	// start program
	srv := startServer(r, appConf.Port)

	// receive quit signal, ready to exit
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Println("Shutdown Server ...")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// startServer 自定义http配置
func startServer(router *gin.Engine, port int) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}

// fileServer 文件服务
func fileServer(port int) {
	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	filePath, _ := os.Getwd()
	mux.Handle("/", http.FileServer(http.Dir(filePath)))
	// 使用基本认证保护文件下载路由
	authMux := basicAuth(mux)

	// 启动 HTTP 服务器
	//log.Printf("Starting file server at :%d", config.GetConfig().Application.Port+1)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), authMux)
	if err != nil {
		log.Fatalf("file server listen: %s\n", err)
	}
	return
}

// basicAuth 认证检查
func basicAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != bootstrap.Username || pass != bootstrap.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
