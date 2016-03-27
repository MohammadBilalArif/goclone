package main

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func sendTars(archives chan io.Reader) {
	for tar := range archives {
		data, _ := ioutil.ReadAll(tar)

		log.Println("Received Tar File:", len(data), "bytes")

		for _, cl := range clients {
			log.Println("Sending to client", cl.Conn.RemoteAddr())

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
	}
}
