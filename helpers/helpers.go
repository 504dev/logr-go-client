package helpers

import (
	"math/rand"
	"sync"
	"time"
)

// rnd is guarded by rndMu: rand.New returns a generator that is not safe for
// concurrent use, unlike the top-level functions of math/rand. RandString is
// called from LogPackage.Chunkify on every log record, and applications log
// from several goroutines at once.
var (
	rndMu sync.Mutex
	rnd   = rand.New(rand.NewSource(time.Now().Unix()))
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)

	rndMu.Lock()
	for i := range b {
		b[i] = letterBytes[rnd.Intn(len(letterBytes))]
	}
	rndMu.Unlock()

	return string(b)
}

func ChunkifyBytes(input []byte, length int) [][]byte {
	result := make([][]byte, 0, 2)
	for i := 0; i < len(input); i += length {
		end := i + length
		if end > len(input) {
			end = len(input)
		}
		result = append(result, input[i:end])
	}
	return result
}

func AddSpacesToMakeMultipleOfN(inputBytes []byte, n int) []byte {
	spacesToAdd := n - (len(inputBytes) % n) // Calculate the number of spaces to add
	if spacesToAdd != n {
		for i := 0; i < spacesToAdd; i++ {
			inputBytes = append(inputBytes, ' ') // Add spaces to the byte slice
		}
	}
	return inputBytes
}
