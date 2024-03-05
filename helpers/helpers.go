package helpers

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rnd.Intn(len(letterBytes))]
	}
	return string(b)
}

func ChunkifyString(input []byte, length int) [][]byte {
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
