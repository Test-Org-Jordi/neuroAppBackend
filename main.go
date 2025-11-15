package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/gin-gonic/gin"
	"neuro.app.jordi/internal/api"
	"neuro.app.jordi/internal/shared/mysql"
)

type Config struct {
	Port string `envconfig:"PORT" default:"8080"`
	Env  string `envconfig:"ENV" default:"development"`
}

func main() {
	ctx := context.Background()
	logger := sentry.NewLogger(ctx)

	// Or inline using WithCtx()
	newCtx := context.Background()
	// WithCtx() does not modify the original context attached on the logger.
	logger.Info().WithCtx(newCtx).Emit("context passed")

	// You can use the logger like [fmt.Print]
	logger.Info().Emit("Hello ", "world!")
	// Or like [fmt.Printf]
	logger.Info().Emitf("Hello %v!", "world")
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://9dd2a33646f1f50da7e3af6956925028@o4510365181935616.ingest.de.sentry.io/4510368658489424",
		Environment:      "prod",
		TracesSampleRate: 1.0, // captura todo
		EnableLogs:       true,
		EnableTracing:    true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(2 * time.Second)

	db, err := mysql.NewMySQL()
	if err != nil {
		log.Fatal(err)
	}

	router := api.NewApp(db).SetupRouter()
	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pongg"})
		})
	}
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == "" || port == ":" {
		port = ":8401"
	}
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	log.Printf("Server listening on port: " + port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Listen error: %s\n", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %s", err)
	}
	log.Println("Server exited cleanly")
}
