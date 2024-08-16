package main

import (
	"encoding/json"
	"github.com/504dev/logr-go-client/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogPackageChunks_isComplete(t *testing.T) {
	tests := []struct {
		name     string
		chunks   types.LogPackageChunks
		expected bool
	}{
		{
			name: "Complete chunks",
			chunks: types.LogPackageChunks{
				&types.LogPackage{PlainLog: []byte("chunk1")},
				&types.LogPackage{PlainLog: []byte("chunk2")},
			},
			expected: true,
		},
		{
			name: "Incomplete chunks",
			chunks: types.LogPackageChunks{
				&types.LogPackage{PlainLog: []byte("chunk1")},
				nil,
				&types.LogPackage{PlainLog: []byte("chunk3")},
			},
			expected: false,
		},
		{
			name:     "Empty chunks",
			chunks:   types.LogPackageChunks{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := tt.chunks.Joined()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogPackageChunks_Joined(t *testing.T) {
	tests := []struct {
		name             string
		chunks           types.LogPackageChunks
		expectedComplete bool
		expectedJoined   *types.LogPackage
	}{
		{
			name: "Join plain logs",
			chunks: types.LogPackageChunks{
				&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 0, N: 2}, PlainLog: []byte("chunk1")},
				&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 1, N: 2}, PlainLog: []byte("chunk2")},
			},
			expectedComplete: true,
			expectedJoined: &types.LogPackage{
				Chunk:    &types.ChunkInfo{Uid: "test", I: 0, N: 2},
				PlainLog: []byte("chunk1chunk2"),
			},
		},
		{
			name: "Join cipher logs",
			chunks: types.LogPackageChunks{
				&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 0, N: 2}, CipherLog: []byte("encrypted1")},
				&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 1, N: 2}, CipherLog: []byte("encrypted2")},
			},
			expectedComplete: true,
			expectedJoined: &types.LogPackage{
				Chunk:     &types.ChunkInfo{Uid: "test", I: 0, N: 2},
				CipherLog: []byte("encrypted1encrypted2"),
			},
		},
		{
			name: "Incomplete chunks",
			chunks: types.LogPackageChunks{
				&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 0, N: 2}, PlainLog: []byte("chunk1")},
				nil,
			},
			expectedComplete: false,
			expectedJoined:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete, joined := tt.chunks.Joined()
			assert.Equal(t, tt.expectedComplete, complete)
			if tt.expectedJoined != nil {
				assert.Equal(t, tt.expectedJoined.Chunk, joined.Chunk)
				if tt.expectedJoined.PlainLog != nil {
					assert.Equal(t, tt.expectedJoined.PlainLog, joined.PlainLog)
				} else {
					assert.Equal(t, tt.expectedJoined.CipherLog, joined.CipherLog)
				}
			} else {
				assert.Nil(t, joined)
			}
		})
	}
}

func TestLogPackageChunks_Marshal(t *testing.T) {
	chunks := types.LogPackageChunks{
		&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 0, N: 2}, PlainLog: []byte("chunk1")},
		&types.LogPackage{Chunk: &types.ChunkInfo{Uid: "test", I: 1, N: 2}, PlainLog: []byte("chunk2")},
	}

	marshaled, err := chunks.Marshal()
	assert.NoError(t, err)
	assert.Len(t, marshaled, 2)

	// Проверяем, что каждый элемент можно успешно демаршализовать обратно в types.LogPackage
	for i, marshaledChunk := range marshaled {
		var unmarshaled types.LogPackage
		err = json.Unmarshal(marshaledChunk, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, chunks[i].Chunk, unmarshaled.Chunk)
		assert.Equal(t, chunks[i].PlainLog, unmarshaled.PlainLog)
	}
}
