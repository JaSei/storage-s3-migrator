package main

import (
	"math/rand"
	"time"

	"github.com/JaSei/pathutil-go"
)

type byteCombination struct {
	firstByte  byteFolder
	secondByte byteFolder
	thirdByte  byteFolder
}

type byteCombinations []byteCombination

func (shard byteRange) visitTree(baseDir pathutil.Path, visitFunc func(pathutil.Path)) {
	var i uint
	comb := make(byteCombinations, shard.length()*256*256)
	shard.lister(func(_ int, firstByte byteFolder) {
		byteRange{0, 255}.lister(func(_ int, secondByte byteFolder) {
			byteRange{0, 255}.lister(func(_ int, thirdByte byteFolder) {
				comb[i] = byteCombination{firstByte, secondByte, thirdByte}
				i++
			})
		})
	})

	for _, bc := range comb.shuffle() {
		visitFunc(bc.path(baseDir))
	}
}

func (slice byteCombinations) shuffle() byteCombinations {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for i := len(slice) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

func (comb byteCombination) path(dir pathutil.Path) pathutil.Path {
	path, _ := pathutil.New(dir.String(), comb.firstByte.hex(), comb.secondByte.hex(), comb.thirdByte.hex())
	return path
}
