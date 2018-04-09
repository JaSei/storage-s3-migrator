package main

import (
	"strings"
	"sync"
	"time"

	"github.com/JaSei/pathutil-go"
	"github.com/alecthomas/kingpin"
	"github.com/avast/retry-go"
	"github.com/avast/storage-s3-migrator/s3"
	"github.com/pkg/errors"
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
	debug              = kingpin.Flag("debug", "log debug").Bool()
)

func main() {
	kingpin.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

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
	for i, shardDir := range shards {
		wg.Add(1)
		go func(id int, shardDir byteRange) {
			defer wg.Done()

			s3cli := newClient()
			log.Info(s3cli)

			shardDir.visitTree(dirPath, func(dir pathutil.Path) {
				dirStat := stat{}
				defer func() { statChannel <- dirStat }()

				if !dir.IsDir() {
					log.Debugf("Skip %s", dir)
					return
				}

				dir.Visit(func(path pathutil.Path) {
					var size int64
					startUploadTime := time.Now()

					exists, err := s3cli.ExistsObject(path)
					if err != nil {
						log.Error(err)
					}

					if exists {
						dirStat.exists++
						log.Infof("Path %s already inserted", path)
					} else {
						err := retry.Do(func() error {
							size, err = s3cli.UploadObject(path)
							return err
						},
							retry.RetryIf(func(err error) bool {
								return err.Error() != "409 Conflict"
							}),
							retry.OnRetry(func(n uint, err error) {
								log.Debugf("Retry %d: %s", n, err.Error())
							}),
						)

						if err != nil {
							if strings.Contains(err.Error(), "409 Conflict") {
								dirStat.exists++
								log.Info(errors.Wrap(err, path.String()))
							} else {
								dirStat.err++
								log.Error(errors.Wrap(err, path.String()))
							}
						} else {
							dirStat.ok++
							log.Info(path)
						}
					}

					duration := time.Since(startUploadTime)
					dirStat.duration += duration
					dirStat.size += size

				}, pathutil.VisitOpt{})
			})

		}(i, shardDir)
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
