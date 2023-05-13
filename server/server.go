package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type RateLimiterHandler[T, TE any] func(ctx context.Context, request T) (response TE)

func AddHandler[T, TE any](pattern string, handler RateLimiterHandler[T, TE], timeout time.Duration, logger *log.Logger) {
	http.HandleFunc(pattern, handleRequest[T, TE](handler, timeout, logger))
	logger.Printf("Registered handler to path: `%s`", pattern)
}

func Listen(address string, logger *log.Logger) {
	if err := http.ListenAndServe(address, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal(err)
	}
}

func LoggerMiddleware[T, TE any](
	handler RateLimiterHandler[T, TE],
	logger *log.Logger,
) RateLimiterHandler[T, TE] {
	return func(ctx context.Context, in T) TE {
		start := time.Now()
		logger.Printf("received request with input: `%v`", in)
		res := handler(ctx, in)
		defer logger.Printf("request processed in `%v` result: `%v`", time.Now().Sub(start), res)
		return res
	}
}

func handleRequest[T, TE any](
	handler RateLimiterHandler[T, TE],
	timeout time.Duration,
	logger *log.Logger,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		var in T

		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			logger.Printf("Error occurred trying to parse the request. Error: `%v`", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := handler(ctx, in)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			logger.Printf("Error occurred trying to parse the response. Error: `%v`", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
