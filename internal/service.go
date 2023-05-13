package internal

import (
	"context"
	"log"
	"sync"
	"time"

	"rate-limiter/config"
	"rate-limiter/internal/models"
	"rate-limiter/internal/utils"
)

type RateLimiterService struct {
	logger        *log.Logger
	threshold     int
	ttl           time.Duration
	numOfWorkers  int
	pollTTLTicker *time.Ticker
	urlCounter    utils.Counter[string]
	ttlQueue      utils.ConditionalQueue[*models.UrlTimestamp]
	done          chan bool
	wg            sync.WaitGroup
}

func NewRateLimiterService(
	config *config.RateLimiterServiceConfig,
	counter *utils.Counter[string],
	ttlQ *utils.ConditionalQueue[*models.UrlTimestamp],
	logger *log.Logger,
) *RateLimiterService {
	return &RateLimiterService{
		threshold:     config.Threshold,
		ttl:           config.TTL,
		logger:        logger,
		numOfWorkers:  config.NumOfWorkers,
		urlCounter:    *counter,
		pollTTLTicker: time.NewTicker(config.PollTTLTickerInterval),
		ttlQueue:      *ttlQ,
		done:          make(chan bool),
		wg:            sync.WaitGroup{},
	}
}

// Start starts the goroutines that poll the ttl sorted queue,
// and remove from memory urls that are past their ttl /**
func (r *RateLimiterService) Start() {
	r.pollTTL()
}

// Handle is the main entry point for report calls.
// This method holds all business logic.
// It enqueues the timestamp for the url if necessary in a different goroutine as it shouldn't hold the client
func (r *RateLimiterService) Handle(_ context.Context, url string) bool {
	count, key := r.urlCounter.Report(url)
	if count >= r.threshold {
		return true
	}

	if count == 0 {
		go r.ttlQueue.Enqueue(&models.UrlTimestamp{
			Time: time.Now(),
			Url:  key,
		})
	}

	return false
}

func (r *RateLimiterService) pollTTLWorker() {
	if t, ok := r.ttlQueue.DequeueIf(func(top *models.UrlTimestamp) bool {
		return time.Now().After(top.Add(r.ttl))
	}); ok {
		r.urlCounter.Reset(t.Url)
		r.pollTTLWorker() // shouldn't wait the interval in case there are more past their TTL
	}
}

func (r *RateLimiterService) pollTTL() {
	for i := 0; i < r.numOfWorkers; i++ {
		r.wg.Add(1)
		go func() {
			for {
				select {
				case <-r.done:
					r.logger.Print("Stopped polling the ttl queue")
					r.wg.Done()
					return
				case <-r.pollTTLTicker.C:
					r.pollTTLWorker()
				}
			}
		}()
	}
}

// Stop being called when program wants to end gracefully.
// Even though most of the things it does should be managed by GC it ensures no memory leaks or open processes
// /**
func (r *RateLimiterService) Stop() {
	r.logger.Print("Received stop call. Shutting down")
	for i := 0; i < r.numOfWorkers; i++ {
		r.done <- true
	}
	r.ttlQueue.Drain()
	r.urlCounter.Drain()
	r.wg.Wait()
}
