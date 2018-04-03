package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
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
	return fmt.Sprintf("Uploaded objects: %d (%d fail), Total size: %s, Speed: %s/s", s.ok, s.err, humanize.Bytes(uint64(s.size)), humanize.Bytes(uint64(float64(s.size)/s.duration.Seconds())))
}
