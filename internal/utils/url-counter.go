package utils

import (
	"hash"
	"log"
	"sync"

	"github.com/puzpuzpuz/xsync"
)

type Counter[T any] interface {
	Report(key T) (int, uint32)
	Reset(hashedKey uint32)
	Drain()
}

type ConcurrentHashCounter[T any] struct {
	xsync.MapOf[uint32, uint32]
	hash           hash.Hash32
	logger         *log.Logger
	lock           *sync.Mutex
	conversionFunc ConvertToBytesFunc[T]
}

type ConvertToBytesFunc[T any] func(T) []byte

func NewConcurrentHashCounter[T any](
	hash hash.Hash32,
	convertFunc ConvertToBytesFunc[T],
	logger *log.Logger,
) Counter[T] {
	return &ConcurrentHashCounter[T]{
		MapOf:          *xsync.NewIntegerMapOf[uint32, uint32](),
		hash:           hash,
		logger:         logger,
		lock:           &sync.Mutex{},
		conversionFunc: convertFunc,
	}
}

func (u *ConcurrentHashCounter[T]) Report(key T) (int, uint32) {
	_, err := u.hash.Write(u.conversionFunc(key))
	if err != nil {
		u.logger.Printf("Failed to hash the key `%s`", key)
		return -1, 0
	}

	hashedKey := u.hash.Sum32()
	u.hash.Reset()

	u.lock.Lock()
	defer u.lock.Unlock()

	count, _ := u.Load(hashedKey)
	u.Store(hashedKey, count+1)

	return int(count), hashedKey
}

func (u *ConcurrentHashCounter[T]) Reset(key uint32) {
	u.MapOf.Delete(key)
}

func (u *ConcurrentHashCounter[T]) Drain() {
	u.Range(func(key uint32, _ uint32) bool {
		u.MapOf.Delete(key)
		if u.Size() == 0 {
			return false
		}
		return true
	})
}
