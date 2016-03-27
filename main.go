package main

import "flag"

func main() {
	clientMode := flag.Bool("client", false, "runs the tool in client mode")
	flag.BoolVar(clientMode, "c", false, "")

	workDir := flag.String("directory", ".", "tool working directory")
	flag.StringVar(workDir, "d", ".", "")

	serverAddr := flag.String("server", "", "ip:port to connect to in client mode")
	flag.StringVar(serverAddr, "s", "", "")

	flag.Parse()

	if *clientMode == true {
		runClient(*workDir, *serverAddr)
	} else {
		runServer(*workDir)
	}
}
