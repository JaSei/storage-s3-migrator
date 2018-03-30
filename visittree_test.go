package main

import (
	"testing"

	"github.com/JaSei/pathutil-go"
	"github.com/stretchr/testify/assert"
)

func TestVisit(t *testing.T) {
	dir, err := pathutil.NewTempDir(pathutil.TempOpt{})
	assert.NoError(t, err)

	i := 0
	byteRange{0, 0}.visitTree(dir, func(path pathutil.Path) {
		i++
	})

	assert.Equal(t, 256*256, i)
}

func TestShuffle(t *testing.T) {
	origin := byteCombinations{byteCombination{1, 2, 3}, byteCombination{2, 3, 4}, byteCombination{3, 4, 5}, byteCombination{4, 5, 6}}
	before := make(byteCombinations, len(origin))
	copy(before, origin)

	after := before.shuffle()
	t.Log(after)
	assert.Equal(t, before, after)
	assert.NotEqual(t, origin, before)
}
