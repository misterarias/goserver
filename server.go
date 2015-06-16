package main

import (
	"./src/goserver"
	"os"
)

func main() {
	port := os.Args[2]

	baseDir := os.Getenv("BASEDIR")
	if baseDir == "" {
		baseDir = "."
	}

	var cmdName string = "watch"
	var cmdArgs = []string{"ls", "-lah"}

	server := goserver.New("Main", cmdName, cmdArgs, port)
	server.NewServer()
}
