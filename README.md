# Rate Limiter

## Introduction

Rate Limiter is an implementation of a web service that limits the rate of requests to a specific url.
To enable maximum performance and minimum memory consumption the implementation uses two concurrent data structures
that together provide all functionalities in O(1).

The idea is to differentiate between client request needs and data maintenance to return the result to the client ASAP,
and maintain the data structures in offline.
---

## Data Structures
* `ConcurrentHashCounter[T]` - a concurrent-safe generic dictionary uint32 --> uint32 that hashes a string URL using the 
[fnv32a](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function)
algorithm (quick, consistent and good uniform distribution). The DS was tweaked to have the `Report` functionality 
(load + store) concurrent-safe. When the web server stops it drains itself to avoid memory leaks.
* `ConcurrentConditionalQueue[T, TE]` - a sorted concurrent-safe generic doubly linked list to hold 
the timestamp and the url hash-key such that, when TTL fires the hash-key could be used to go and remove 
the entry from the URL counter DS. The DS was tweaked to have the `Dequeue` operation conditional and still concurrent-safe.
When the web server stops it drains itself to avoid memory leaks.
---

## Rate Limiter Service
The `RateLimiterService` is the class to hold business logic and operate the datastructures.
configuration for it are as follows:
* `ttl` - time interval for any url. Received from the program arguments as MS.
* `threshold` - the amount of requests to allow in the provided `ttl`. Received from the program arguments.
* `numOfWorkers` - the amount of goroutines to run over the TTL queue. Hard-coded 3 for now.
* `pollTTLTickerInterval` - The interval between checks for if the minimum TTL had past its time.
---

## The Server
The `server.go` file holds static functions that wrap the http package such that as a client 
you can easily and generically add handlers to different patterns to open new endpoints with 
different `Request` / `Response` objects. I also added a `LoggerMiddleware` function (shouldn't be used in production)
to have robust logging for every request with its input, output and process duration.
---

## Collisions
It's inherited in the design that the occurrence is very rare.
For a collision to happen it requires, in average, more than 2^31 distinct urls in the DS, all are not past their TTL.
If a collision does happen, the result would be inaccuracy in the response (sum of both url reports).
The system will keep operating as usual except for that. Thus, it's not being handled.

---
## Final Note
The project as a whole is aiming to keep types loosely coupled by using dependency injection to have
dependency inversion principle in place. The mindset was - if something can be generic then it will.

Theoretically the projects efficiency can be improved by implementing an almost lock-free doubly linked list
by atomically swap the pointers (doubly) to the last element when adding a new element `Enqueue`, 
and atomically swap the pointers to the first element when popping the first element `Dequeue`