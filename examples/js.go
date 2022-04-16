//go:build js
// +build js

package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"githun.com/arion-dsh/aowoo"
)

type jsfile struct {
	*bytes.Reader
}

func (f *jsfile) Close() error {
	return nil
}

func openFile(path string) (io.ReadCloser, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join("assets", path)
	}
	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	f := &jsfile{bytes.NewReader(body)}
	return f, nil
}

func main() {

	f, err := openFile("2.wav")
	if err != nil {
		panic("no sound file")
	}
	aowoo.Open(44100, 16, 2)
	s, _ := aowoo.NewSource(f, 1, false)
	aowoo.Play(s)

	ch := make(chan int, 1)
	<-ch

}
