package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func sendAllFiles(workDir string, cl client) {
	archives := make(chan io.Reader, 1)

	wg := sync.WaitGroup{}
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

		dir, name := path.Split(p)

		hdr := fileHeader{
			Name:         name,
			Path:         dir,
			Mode:         int64(info.Mode()),
			LastModified: info.ModTime(),
		}

		c <- hdr

		return nil

	})

	wg.Wait()

	close(c)

	tar := <-archives

	data, _ := ioutil.ReadAll(tar)

	log.Println("Sending All Files to New Client")

	go func(c client) {
		buf := make([]byte, 4)
		binary.Write(c.Conn, binary.LittleEndian, uint32(len(data)))
		c.Conn.Write(data)
		c.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		c.Conn.Read(buf)
		if string(buf) != "ok  " {
			c.Conn.Close()
			delete(clients, c.Conn)
		}
	}(cl)
}

func runClient(workDir string, serverAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalln("[!] unable to dial server ", serverAddr, err)
	}

	defer conn.Close()

	var size int32

	for {
		if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			log.Println("[!] error reading from net connection", err)
			break
		}

		out := bytes.Buffer{}
		buf := make([]byte, 4096)
		var bytesRead int32

		for bytesRead < size {
			if size-bytesRead < 4096 {
				n, err := conn.Read(buf[:size-bytesRead])

				if err == io.EOF {
					break
				}
				if err != nil {
					log.Println("[!] unable to read tar archive from net connection", err)
					break
				}
				out.Write(buf[:n])

				bytesRead += int32(n)
			} else {
				n, err := conn.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Println("[!] unable to read tar archive from net connection", err)
					break
				}
				out.Write(buf[:n])

				bytesRead += int32(n)
			}
		}

		if len(out.Bytes()) > 0 {
			expandFiles(workDir, &out)

			io.WriteString(conn, "ok  ")
		}
	}
}
