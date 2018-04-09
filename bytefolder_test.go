package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHex(t *testing.T) {
	var bf byteFolder = 255
	assert.Equal(t, "FF", bf.hex())
}
