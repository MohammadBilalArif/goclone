package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func watchDir(archives chan io.Reader, workDir string) {
	timer := time.Tick(5 * time.Second)
	wg := sync.WaitGroup{}

	for _ = range timer {
		buf := bytes.Buffer{}
		c := make(chan fileHeader, 1)

		go archiveFiles(workDir, c, &buf, archives)

		// every five seconds, check all files
		filepath.Walk(workDir, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			p = strings.TrimPrefix(p, path.Clean(workDir)+"/")

			wg.Add(1)
			defer wg.Done()

			if hdr, ok := fileStats[p]; !ok {
				dir, name := path.Split(p)

				fileStats[p] = fileHeader{
					Name:         name,
					Path:         dir,
					Mode:         int64(info.Mode()),
					LastModified: info.ModTime(),
				}

				c <- fileStats[p]
			} else {
				if info.ModTime().After(fileStats[p].LastModified) {
					hdr.LastModified = info.ModTime()
					fileStats[p] = hdr

					log.Println("[*] File Modified: ", p)

					c <- hdr
				}
			}

			return nil

		})

		wg.Wait()

		close(c)
	}
}
