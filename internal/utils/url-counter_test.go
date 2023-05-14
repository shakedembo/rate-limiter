package utils

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type UrlCounterTestSuite struct {
	suite.Suite
	counter Counter[string]
}

func (u *UrlCounterTestSuite) SetupTest() {
	logger := log.New(os.Stdout, "rater-limiter", log.LstdFlags|log.Lshortfile)

	convertToBytesFunc := func(url string) []byte { return []byte(url) }
	hash := NewFnv32aHashProvider(convertToBytesFunc, logger)
	u.counter = NewConcurrentHashCounter(hash, logger)
}

func (u *UrlCounterTestSuite) TestUrlCounterConcurrency() {
	wg := sync.WaitGroup{}
	concurrency := 10000000
	pass := false

	for i := 0; i <= concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, _ := u.counter.Report("google.com")
			if val == concurrency {
				pass = true
			}
		}()
	}

	time.Sleep(2 * time.Second)
	wg.Wait()

	u.Require().True(pass)
}

func TestUrlCounterTestSuite(t *testing.T) {
	suite.Run(t, new(UrlCounterTestSuite))
}
