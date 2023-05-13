package models

import (
	"time"
)

type UrlTimestamp struct {
	time.Time
	Url UrlHash
}

type UrlHash = uint32
