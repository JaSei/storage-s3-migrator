package main

import (
	"sync"
	"time"

	"github.com/JaSei/pathutil-go"
	"github.com/alecthomas/kingpin"
	"github.com/avast/primary-storage-migrator/s3"
	log "github.com/sirupsen/logrus"
)

var (
	dir                = kingpin.Arg("dir", "source directory").Required().ExistingDir()
	concurrent         = kingpin.Flag("concurrent", "count of concurent uploader").Default("8").Uint8()
	endpoint           = kingpin.Flag("endpoint", "endpoint hostname").Required().String()
	namespace          = kingpin.Flag("namespace", "endpoint namespace").Required().String()
	user               = kingpin.Flag("user", "username").Required().String()
	pass               = kingpin.Flag("pass", "password").Required().String()
	customLastModified = kingpin.Flag("custom-last-modifed", "set x-amz-meta-Last-Modified header with last modification time of source file").Bool()
)

func main() {
	kingpin.Parse()
	log.SetLevel(log.DebugLevel)

	dirPath, err := pathutil.New(*dir)
	if err != nil {
		log.Fatal(err)
	}

	shards, err := shardHex(0xFF, (byteFolder)(*concurrent))
	if err != nil {
		log.Fatal(err)
	}

	statChannel, totalStatsChannel := makeStatChannels(*concurrent)
	go statsHandler(statChannel, totalStatsChannel)

	wg := sync.WaitGroup{}
	for i, shardFolder := range shards {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			s3cli := newClient()
			log.Info(s3cli)

			shardFolder.visitTree(dirPath, func(dir pathutil.Path) {
				dirStat := stat{}
				defer func() { statChannel <- dirStat }()

				if !dir.IsDir() {
					log.Debugf("Skip %s", dir)
					return
				}

				dir.Visit(func(path pathutil.Path) {
					startUploadTime := time.Now()
					size, err := s3cli.UploadObject(path)
					duration := time.Since(startUploadTime)

					dirStat.duration += duration
					dirStat.size += size

					if err != nil {
						dirStat.err += 1
						log.Error(err)
					} else {
						dirStat.ok += 1
						log.Info(path)
					}

				}, pathutil.VisitOpt{})
			})

		}(i)
	}

	wg.Wait()
	close(statChannel)
	total := <-totalStatsChannel
	log.Info(total.formatStats())
}

func newClient() hs3.Hs3Client {
	cli, err := hs3.New(*endpoint, *namespace, *user, *pass, *customLastModified)
	if err != nil {
		log.Fatal(err)
	}

	return cli
}
