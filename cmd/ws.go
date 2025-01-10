package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"

	"exapp-go/internal/services/ws"
)

var WebSocketCommand = &cli.Command{
	Name:  "ws",
	Usage: "Start WebSocket service",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Value:   "0.0.0.0",
			Usage:   "Service listen address",
			EnvVars: []string{"WS_HOST"},
		},
		&cli.IntFlag{
			Name:    "port",
			Value:   8081,
			Usage:   "Service listen port",
			EnvVars: []string{"WS_PORT"},
		},
		&cli.StringFlag{
			Name:    "mode",
			Value:   "debug",
			Usage:   "Run mode (debug/release)",
			EnvVars: []string{"GIN_MODE"},
		},
	},
	Action: runWebSocketServer,
}

func runWebSocketServer(c *cli.Context) error {
	// Set gin mode
	gin.SetMode(c.String("mode"))

	// Create gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create WebSocket server
	wsServer := ws.NewServer(ctx)
	defer wsServer.Close()

	// Set WebSocket routes
	r.GET("/socket.io/*any", gin.WrapH(wsServer.Handler()))
	r.POST("/socket.io/*any", gin.WrapH(wsServer.Handler()))

	// Set static file service
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// Create HTTP server
	addr := c.String("host") + ":" + c.String("port")
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		// Listen for interrupt signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		// Create shutdown context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
	}()

	// Start server
	log.Printf("Server is running at http://%s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
		return err
	}

	log.Println("Server exited")
	return nil
}
