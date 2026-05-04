# GORedis
GORedis  
GORedis — lightweight in-memory key-value store inspired by Redis, written in Go.

Features:

	* GET / SET / DELETE operations via HTTP API
	* Persistent storage with MessagePack serialization
	* Atomic file writes (temp file + rename)
	* Concurrent access with RWMutex
	* Graceful shutdown with data saving
