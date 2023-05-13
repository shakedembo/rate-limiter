package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"os/signal"
	"time"

	"rate-limiter/config"
	"rate-limiter/internal"
	"rate-limiter/internal/models"
	"rate-limiter/internal/utils"
	"rate-limiter/pkg"
	"rate-limiter/server"
)

var logger = log.New(os.Stdout, "rater-limiter", log.LstdFlags|log.Lshortfile)

func main() {
	logger.Print("Initializing...")
	defer logger.Printf("Bye Bye :)")

	hash := fnv.New32a()
	counter := utils.NewConcurrentHashCounter(hash, func(url string) []byte { return []byte(url) }, logger)
	ttlQ := utils.NewConcurrentConditionalQueue[*models.UrlTimestamp]()
	serviceConfig := config.NewRateLimiterServiceConfig(logger)
	rateLimiterService := internal.NewRateLimiterService(serviceConfig, &counter, &ttlQ, logger)

	rateLimiterService.Start()

	reportPattern := "/report"
	reportTimeout := 2 * time.Second

	server.AddHandler[pkg.Request, pkg.Response](reportPattern,
		//middleware to print request and response with process time. (debug mode)
		server.LoggerMiddleware[pkg.Request, pkg.Response](
			func(ctx context.Context, request pkg.Request) (response pkg.Response) {
				return pkg.Response{
					Block: rateLimiterService.Handle(ctx, request.Url),
				}
			},
			logger),
		reportTimeout,
		logger)

	port := "8080" // os.Getenv("SERVER_PORT")
	go server.Listen(fmt.Sprintf(":%s", port), logger)

	logger.Printf("Listening on port `%s`", port)
	logger.Print("--- Initialization Completed Successfully ---")

	gracefullyQuit(rateLimiterService)
}

func gracefullyQuit(rateLimiterService *internal.RateLimiterService) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rateLimiterService.Stop()
}
