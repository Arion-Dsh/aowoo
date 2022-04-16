//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"log"
	"os"
	"path/filepath"

	"githun.com/arion-dsh/aowoo"
)

func main() {

	path := filepath.Join("assets", "2.wav")
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	aowoo.Open(44100, 16, 2)
	aowoo.NewSource(f, 1, true)

	ch := make(chan int, 1)
	<-ch
}
