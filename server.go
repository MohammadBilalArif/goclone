package main

import (
	"io"
	"log"
	"net"
	"time"
)

type fileHeader struct {
	Name         string
	Path         string
	Mode         int64
	LastModified time.Time
}

type client struct {
	Conn     net.Conn
	LastSeen time.Time
}

var fileStats map[string]fileHeader
var clients map[net.Conn]client

func runServer(workDir string) {

	fileStats = make(map[string]fileHeader)
	clients = make(map[net.Conn]client)

	archives := make(chan io.Reader, 1)

	go watchDir(archives, workDir)
	go sendTars(archives)

	conn, err := net.Listen("tcp", ":3003")
	if err != nil {
		log.Fatalln("[!] unable to listen to server port", err)
	}

	defer conn.Close()

	for {
		c, err := conn.Accept()
		if err != nil {
			log.Println("[-] unable to accept client connection", err)
			continue
		}

		clients[c] = client{
			LastSeen: time.Now(),
			Conn:     c,
		}

		sendAllFiles(workDir, clients[c])
	}
}
