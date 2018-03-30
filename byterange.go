package main

import (
	//	"github.com/JaSei/pathutil-go"
	"github.com/pkg/errors"
	//	log "github.com/sirupsen/logrus"
)

type byteRange struct {
	min byteFolder
	max byteFolder
}

func shardHex(max, count byteFolder) ([]byteRange, error) {
	if count == 0 {
		return []byteRange{}, errors.New("Count must be grather then zero")
	}
	if count > max {
		return []byteRange{}, errors.New("Count is grather then max")
	}

	piece := (int(max) + 1) / int(count)

	byteRanges := make([]byteRange, count)

	for i := 0; i < int(count); i++ {
		byteRanges[i] = byteRange{byteFolder(i * piece), byteFolder((i+1)*piece - 1)}
	}

	if byteRanges[count-1].max != max {
		byteRanges[count-1].max = max
	}

	return byteRanges, nil
}

func (shard byteRange) list() byteFolders {
	hexList := make(byteFolders, shard.length())
	shard.lister(func(i int, bf byteFolder) {
		hexList[i] = bf
	})
	return hexList
}

func (shard byteRange) lister(listerFunc func(int, byteFolder)) {
	for i := int(shard.min); i <= int(shard.max); i++ {
		listerFunc(i-int(shard.min), byteFolder(i))
	}
}

func (shard byteRange) length() uint {
	return (uint(shard.max) - uint(shard.min)) + 1
}

//func (shard byteRange) visitTree(baseDir pathutil.Path, visitFunc func(pathutil.Path)) {
//	for _, firstByte := range shard.list().shuffle() {
//		for _, secondByte := range (byteRange{0, 255}.list().shuffle()) {
//			for _, thirdByte := range (byteRange{0, 255}.list().shuffle()) {
//				dir, _ := pathutil.New(baseDir.String(), firstByte.hex(), secondByte.hex(), thirdByte.hex())
//
//				if !dir.IsDir() {
//					log.Debugf("Skip %s", dir)
//					continue
//				}
//
//				visitFunc(dir)
//			}
//		}
//	}
//
//}
