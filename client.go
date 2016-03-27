package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
)

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
