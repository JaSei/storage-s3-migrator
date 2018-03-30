package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cheggaaa/pb"
)

type stat struct {
	err      int
	ok       int
	size     int64
	duration time.Duration
}

func makeStatChannels(size uint8) (chan stat, chan stat) {
	return make(chan stat, size*10), make(chan stat, 1)
}

const totalDirs = 256 * 256 * 256

func statsHandler(statChannel <-chan stat, totalStatsChannel chan<- stat) {
	progress := pb.Full.Start(totalDirs)
	progress.SetRefreshRate(time.Second)
	progress.SetWriter(os.Stdout)
	var total stat

	//progress.PrependFunc(func(b *uiprogress.Bar) string {
	//	return fmt.Sprintf("Dir %d/%d, %s", b.Current(), totalDirs, total.formatStats())
	//})

	for s := range statChannel {
		progress.Increment()

		total.size += s.size
		total.duration += s.duration
		total.err += s.err
		total.ok += s.ok

		progress.Set("prefix", total.formatStats()+", Dirs: ")
	}

	progress.Finish()

	totalStatsChannel <- total
}

func (s stat) formatStats() string {
	return fmt.Sprintf("Uploaded objects: %d (%d fail), Total size: %dMB", s.ok, s.err, s.size/totalDirs)
}
