package helpers

import (
	"sync"
	"testing"
)

// TestRandStringConcurrent ловит гонку на общем генераторе случайных чисел.
// RandString вызывается из LogPackage.Chunkify, то есть на каждой записи лога,
// а логи в приложении пишут разные горутины одновременно.
//
// Без синхронизации падает под -race.
func TestRandStringConcurrent(t *testing.T) {
	const goroutines = 8
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if got := RandString(6); len(got) != 6 {
					t.Errorf("RandString(6) = %q, want 6 chars", got)
					return
				}
			}
		}()
	}

	wg.Wait()
}
