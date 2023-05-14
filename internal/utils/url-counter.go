package utils

import (
	"log"
	"sync/atomic"

	"github.com/puzpuzpuz/xsync"
)

type Counter[T any] interface {
	Report(key T) (int, uint32)
	Reset(hashedKey uint32)
	Drain()
}

type ConcurrentHashCounter[T any] struct {
	xsync.MapOf[uint32, *uint32]
	hash           HashProvider[T]
	logger         *log.Logger
	conversionFunc ConvertToBytesFunc[T]
}

func NewConcurrentHashCounter[T any](
	hash HashProvider[T],
	logger *log.Logger,
) Counter[T] {
	return &ConcurrentHashCounter[T]{
		MapOf:  *xsync.NewIntegerMapOf[uint32, *uint32](),
		hash:   hash,
		logger: logger,
	}
}

func (u *ConcurrentHashCounter[T]) Report(key T) (int, uint32) {
	hashedKey := u.hash.Get(key)
	var ptr uint32 = 1

	count, exists := u.MapOf.LoadOrStore(hashedKey, &ptr)
	if exists {
		atomic.AddUint32(count, 1)
	}

	return int(*count), hashedKey
}

func (u *ConcurrentHashCounter[T]) Reset(key uint32) {
	u.MapOf.Delete(key)
}

func (u *ConcurrentHashCounter[T]) Drain() {
	u.Range(func(key uint32, _ *uint32) bool {
		u.MapOf.Delete(key)
		if u.Size() == 0 {
			return false
		}
		return true
	})
}
