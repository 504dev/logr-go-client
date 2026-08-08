package main

import (
	"sync"
	"testing"

	logr "github.com/504dev/logr-go-client"
)

// TestLoggerConcurrentWrites checks that a logger can be written to from several
// goroutines, which is how applications use it: workers, HTTP handlers and
// background loops all log through the same instance.
//
// Nothing listens on the UDP address — the transport only needs to be dialable.
func TestLoggerConcurrentWrites(t *testing.T) {
	conf := logr.Config{Udp: "127.0.0.1:65001"}

	logger, err := conf.NewLogger("concurrent-test.log")
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}
	defer logger.Close()

	// Console output would only add noise to the test log.
	logger.Console = false

	const goroutines = 8
	const iterations = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				logger.Info("goroutine", id, "iteration", j)
			}
		}(i)
	}

	wg.Wait()
}
