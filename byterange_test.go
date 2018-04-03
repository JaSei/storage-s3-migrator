package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShardHex(t *testing.T) {
	t.Run("00-FF, 1 => 00-FF", func(t *testing.T) {
		shards, err := shardHex(0xFF, 1)
		assert.NoError(t, err)
		assert.Equal(t, []byteRange{{0, 0xff}}, shards)
	})

	t.Run("00-FF, 2 => 00-7F, 80-FF", func(t *testing.T) {
		shards, err := shardHex(0xFF, 2)
		assert.NoError(t, err)
		assert.Equal(t, []byteRange{{0, 0x7f}, {0x80, 0xff}}, shards)

	})

	t.Run("00-FF, 3 => 00-54, 55-A9, AA-FF", func(t *testing.T) {

		shards, err := shardHex(0xFF, 3)
		assert.NoError(t, err)
		assert.Equal(t, []byteRange{{0, 0x54}, {0x55, 0xA9}, {0xAA, 0xFF}}, shards)
	})

	t.Run("00-FF, 8 => 00-1F, 20-3F, , ,...", func(t *testing.T) {

		shards, err := shardHex(0xFF, 8)
		assert.NoError(t, err)
		assert.Equal(t, []byteRange{{0, 0x1F}, {0x20, 0x3F}, {0x40, 0x5F}, {0x60, 0x7F}, {0x80, 0x9F}, {0xA0, 0xBF}, {0xC0, 0xDF}, {0xE0, 0xFF}}, shards)
	})

	t.Run("00-FF, 255 => 00, 01, ..., FF", func(t *testing.T) {
		shards, err := shardHex(0xFF, 255)
		assert.NoError(t, err)
		assert.Equal(t, byteRange{0x0, 0x0}, shards[0])
		assert.Equal(t, byteRange{0xFE, 0xFF}, shards[254])
	})

	t.Run("Error", func(t *testing.T) {
		_, err := shardHex(128, 129)
		assert.Error(t, err)

		_, err = shardHex(0xFF, 0)
		assert.Error(t, err)
	})
}

func TestHexList(t *testing.T) {
	shards, err := shardHex(0xFF, 32)
	assert.NoError(t, err)
	assert.Equal(t, byteFolders{0, 1, 2, 3, 4, 5, 6, 7}, shards[0].list())
	assert.Equal(t, byteFolders{248, 249, 250, 251, 252, 253, 254, 255}, shards[31].list())

	shards, err = shardHex(0xFF, 255)
	assert.NoError(t, err)
	assert.Equal(t, byteFolders{0}, shards[0].list())
	assert.Equal(t, byteFolders{254, 255}, shards[254].list())
}
