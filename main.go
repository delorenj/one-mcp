package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"one-mcp/backend/api/middleware"
	"one-mcp/backend/api/route"
	"one-mcp/backend/common"
	"one-mcp/backend/common/i18n"
	"one-mcp/backend/library/proxy"
	"one-mcp/backend/model"

	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist
var buildFS embed.FS

//go:embed frontend/dist/index.html
var indexPage []byte

func main() {
	flag.Parse()
	if *common.PrintVersion {
		println(common.Version)
		os.Exit(0)
	}
	if *common.PrintHelpFlag {
		common.PrintHelp()
		os.Exit(0)
	}
	common.SetupGinLog()
	common.SysLog("One MCP Backend" + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	// Initialize Redis
	err := common.InitRedisClient()
	if err != nil {
		common.FatalLog(err)
	}
	// Initialize SQL Database
	err = model.InitDB()
	if err != nil {
		common.FatalLog(err)
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog(err)
		}
	}()

	// Initialize i18n
	localesPath := "./backend/locales"
	// In Docker environment, try absolute path if relative path fails
	err = i18n.Init(localesPath)
	if err != nil {
		localesPath = "/backend/locales"
		err = i18n.Init(localesPath)
	}
	if err != nil {
		common.SysError("Failed to initialize i18n: " + err.Error())
		// Continue without i18n rather than failing completely
	} else {
		common.SysLog("i18n initialized successfully from: " + localesPath)
	}

	// Seed default services
	// if err := model.SeedDefaultServices(); err != nil {
	// 	common.SysError(fmt.Sprintf("Failed to seed default services: %v", err))
	// 	// Depending on severity, might os.Exit(1) or just log
	// }

	// Initialize service manager
	serviceManager := proxy.GetServiceManager()
	go func() {
		if err := serviceManager.Initialize(context.Background()); err != nil {
			common.SysLog("Failed to initialize service manager: " + err.Error())
		} else {
			common.SysLog("Service manager initialized successfully")
		}
	}()

	// Initialize HTTP server
	engine := gin.Default()
	//engine.Use(gzip.Gzip(gzip.DefaultCompression))
	engine.Use(middleware.CORS())

	route.SetRouter(engine, buildFS, indexPage)

	port := strconv.Itoa(*common.Port)
	common.SysLog("Server listening on port: " + port)

	// Create custom http.Server with no IdleTimeout (0) to support long-lived connections
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      engine,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  0,
	}

	// Setup graceful shutdown
	setupGracefulShutdown(srv)

	// Start HTTP server
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("failed to start server: " + err.Error())
	}
}

// setupGracefulShutdown registers signal handlers to ensure clean shutdown
func setupGracefulShutdown(srv *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		common.SysLog("Shutting down...")

		// 关闭服务管理器
		serviceManager := proxy.GetServiceManager()
		if err := serviceManager.Shutdown(context.Background()); err != nil {
			common.SysLog("Error shutting down service manager: " + err.Error())
		} else {
			common.SysLog("Service manager shut down successfully")
		}

		// Gracefully shut down HTTP server
		if err := srv.Shutdown(context.Background()); err != nil {
			common.SysLog("HTTP server Shutdown: " + err.Error())
		}

		// 关闭其他资源...

		os.Exit(0)
	}()
}
