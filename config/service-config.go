package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

/*
* Hard coded configurations that should be put externally in production env.
 */
const (
	numOfWorkers          = 3
	pollTTLTickerInterval = 10 * time.Millisecond
)

type RateLimiterServiceConfig struct {
	Threshold             int
	NumOfWorkers          int
	TTL                   time.Duration
	PollTTLTickerInterval time.Duration
}

func NewRateLimiterServiceConfig(logger *log.Logger) *RateLimiterServiceConfig {
	threshold, ttl := getUserArgs(logger)

	return &RateLimiterServiceConfig{
		Threshold:             threshold,
		NumOfWorkers:          numOfWorkers,
		TTL:                   ttl,
		PollTTLTickerInterval: pollTTLTickerInterval,
	}
}

func getUserArgs(logger *log.Logger) (int, time.Duration) {
	if len(os.Args) != 3 {
		logger.Fatalf(
			"Wrong number of arguments was received."+
				"Expected `2`, received `%d`. Shutting down...",
			len(os.Args)-1)
	}

	threshold, err := strconv.ParseInt(os.Args[1], 0, 0)
	if err != nil {
		logger.Fatalf("The provided threshold is not in the correct form. Error: `%v`", err)
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%sms", os.Args[2]))
	if err != nil {
		logger.Fatalf("The provided ttl is not in the correct form. Error: `%v`", err)
	}

	return int(threshold), duration
}
